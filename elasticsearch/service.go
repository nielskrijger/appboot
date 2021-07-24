package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/estransport"
	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/utils"
)

var (
	errMissingConfig    = errors.New("missing elasticsearch configuration")
	errMissingAddresses = errors.New("config \"elasticsearch.addresses\" is required")
)

type Service struct {
	*elasticsearch7.Client
	*elasticsearch7.Config
}

func (s *Service) Name() string {
	return "elasticsearch"
}

type ClusterInfo struct {
	ClusterName string `json:"cluster_name"`
}

func (s *Service) Configure(ctx *goboot.AppContext) error {
	// unmarshal config and set defaults
	s.Config = &elasticsearch7.Config{}

	if !ctx.Config.InConfig("elasticsearch") {
		return errMissingConfig
	}

	if err := ctx.Config.Sub("elasticsearch").UnmarshalExact(&s.Config); err != nil {
		return fmt.Errorf("parsing elasticsearch configuration: %w", err)
	}

	if len(s.Config.Addresses) == 0 {
		return errMissingAddresses
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

	return s.testConnectivity(ctx)
}

func (s *Service) testConnectivity(ctx *goboot.AppContext) error {
	// Start client
	es, err := elasticsearch7.NewClient(*s.Config)
	if err != nil {
		return fmt.Errorf("creating elasticsearch client: %w", err)
	}

	res, err := es.Info()
	if err != nil {
		return fmt.Errorf("fetch elasticsearch cluster info: %w", err)
	}
	defer utils.Close(ctx.Log, res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf( // nolint:goerr113
			"expected 200 OK but got %q while retrieving elasticsearch info: %s",
			res.Status(),
			res.Body,
		)
	}

	var info ClusterInfo
	if err = json.NewDecoder(res.Body).Decode(&info); err != nil {
		return fmt.Errorf("decoding cluster info: %w", err)
	}

	ctx.Log.Info().Msgf("Successfully connected to ElasticSearch cluster \"%s\"", info.ClusterName)

	return nil
}

func (s *Service) Init() error {
	return nil
}

func (s *Service) Close() error {
	return nil
}
