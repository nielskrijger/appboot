package esboot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/estransport"
	"github.com/nielskrijger/goboot"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

var errMissingElasticsearchAddresses = errors.New("config \"elasticsearch.addresses\" is required")

const defaultMigrationsIndex = "migrations"

type ESClusterInfo struct {
	ClusterName string `json:"cluster_name"`
}

type Elasticsearch struct {
	Migrations      []*Migration
	MigrationsIndex string

	*elasticsearch7.Client
	*elasticsearch7.Config

	log zerolog.Logger
}

func (s *Elasticsearch) Name() string {
	return "Elasticsearch"
}

func (s *Elasticsearch) Configure(env *goboot.AppEnv) error {
	s.log = env.Log

	// Fetch config from viper. Avoid unmarshal directly into elasticsearch7.Config
	// as it doesn't work with env vars:
	// https://github.com/spf13/viper/issues/761
	s.Config = &elasticsearch7.Config{
		Addresses: env.Config.GetStringSlice("elasticsearch.addresses"),
		Username:  env.Config.GetString("elasticsearch.username"),
		Password:  env.Config.GetString("elasticsearch.password"),
	}

	if len(s.Config.Addresses) == 0 {
		return errMissingElasticsearchAddresses
	}

	if s.MigrationsIndex == "" {
		if env.Config.IsSet("elasticsearch.migrationsIndex") {
			s.MigrationsIndex = env.Config.GetString("elasticsearch.migrationsIndex")
		} else {
			s.MigrationsIndex = defaultMigrationsIndex
		}
	}

	// setup debug logging
	if env.Log.Debug().Enabled() {
		human := env.Config.Get("log.human")
		if human == "true" {
			s.Config.Logger = &estransport.ColorLogger{
				Output:             os.Stdout,
				EnableRequestBody:  true,
				EnableResponseBody: true,
			}
		} else {
			s.Config.Logger = &estransport.JSONLogger{
				Output:             os.Stdout,
				EnableRequestBody:  true,
				EnableResponseBody: true,
			}
		}
	}

	// Start client
	client, err := elasticsearch7.NewClient(*s.Config)
	if err != nil {
		return fmt.Errorf("creating Elasticsearch client: %w", err)
	}

	s.Client = client

	return s.testConnectivity(env)
}

func (s *Elasticsearch) testConnectivity(env *goboot.AppEnv) error {
	res, err := s.Client.Info()
	if err != nil {
		return fmt.Errorf("fetch Elasticsearch cluster info: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			env.Log.Warn().Err(err).Msg("failed to properly close Elasticsearch response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf( // nolint:goerr113
			"expected 200 OK but got %q while retrieving Elasticsearch info: %s",
			res.Status(),
			res.Body,
		)
	}

	var info ESClusterInfo
	if err = json.NewDecoder(res.Body).Decode(&info); err != nil {
		return fmt.Errorf("decoding cluster info: %w", err)
	}

	env.Log.Info().Msgf("successfully connected to Elasticsearch cluster \"%s\"", info.ClusterName)

	return nil
}

// Init runs the Elasticsearch migrations.
func (s *Elasticsearch) Init() error {
	if err := s.Migrate(context.Background()); err != nil {
		return fmt.Errorf("running Elasticsearch migrations: %w", err)
	}

	return nil
}

func (s *Elasticsearch) Close() error {
	return nil
}

// ParseResponse decodes the Elasticsearch response body. The response body may
// contain errors which is why it's advisable to always parse the response even
// you're not interested in the actual body.
//
// If r is nil does not decode non-error response body.
//
// Closes the response body when done.
func (s *Elasticsearch) ParseResponse(res *esapi.Response, v any) (err error) {
	b, err := s.ParseResponseBytes(res)
	if err != nil {
		return err
	}

	if v != nil {
		results := gjson.GetBytes(b, "hits.hits.#._source").Raw
		if err := json.Unmarshal([]byte(results), &v); err != nil {
			return fmt.Errorf("parsing Elasticsearch response body: %w", err)
		}
	}

	return nil
}

// ParseResponseBytes parses the Elasticsearch response body to a byte array.
// The response body may contain errors which is why it's advisable to always
// parse the response even you're not interested in the actual body.
//
// Closes the response body when done.
func (s *Elasticsearch) ParseResponseBytes(res *esapi.Response) ([]byte, error) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.log.Warn().Err(err).Msg("error closing the Elasticsearch response reader")
		}
	}(res.Body)

	if res.IsError() {
		// Print the response status and error information.
		return nil, fmt.Errorf("Elasticsearch error response: %s", res.String())
	}

	result, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Elasticsearch response body: %w", err)
	}

	return result, nil
}
