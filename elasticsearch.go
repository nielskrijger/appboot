package goboot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/estransport"
)

var (
	errMissingElasticsearchConfig    = errors.New("missing elasticsearch configuration")
	errMissingElasticsearchAddresses = errors.New("config \"elasticsearch.addresses\" is required")
)

type ESClusterInfo struct {
	ClusterName string `json:"cluster_name"`
}

type ElasticSearch struct {
	*elasticsearch7.Client
	*elasticsearch7.Config
}

func (s *ElasticSearch) Name() string {
	return "elasticsearch"
}

func (s *ElasticSearch) Configure(ctx *AppContext) error {
	// unmarshal config and set defaults
	s.Config = &elasticsearch7.Config{}

	if !ctx.Config.InConfig("elasticsearch") {
		return errMissingElasticsearchConfig
	}

	if err := ctx.Config.Sub("elasticsearch").Unmarshal(&s.Config); err != nil {
		return fmt.Errorf("parsing elasticsearch configuration: %w", err)
	}

	if len(s.Config.Addresses) == 0 {
		return errMissingElasticsearchAddresses
	}

	// setup debug logging
	if ctx.Log.Debug().Enabled() {
		human := ctx.Config.Get("log.human")
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
		return fmt.Errorf("creating elasticsearch client: %w", err)
	}

	s.Client = client

	return s.testConnectivity(ctx)
}

func (s *ElasticSearch) testConnectivity(ctx *AppContext) error {
	res, err := s.Client.Info()
	if err != nil {
		return fmt.Errorf("fetch elasticsearch cluster info: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			ctx.Log.Warn().Err(err).Msg("failed to properly close elasticsearch response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf( // nolint:goerr113
			"expected 200 OK but got %q while retrieving elasticsearch info: %s",
			res.Status(),
			res.Body,
		)
	}

	var info ESClusterInfo
	if err = json.NewDecoder(res.Body).Decode(&info); err != nil {
		return fmt.Errorf("decoding cluster info: %w", err)
	}

	ctx.Log.Info().Msgf("Successfully connected to ElasticSearch cluster \"%s\"", info.ClusterName)

	return nil
}

func (s *ElasticSearch) Init() error {
	return nil
}

func (s *ElasticSearch) Close() error {
	return nil
}
