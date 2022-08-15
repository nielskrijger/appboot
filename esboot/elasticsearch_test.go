package esboot_test

import (
	"context"
	"github.com/nielskrijger/goboot/esboot"
	"os"
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/stretchr/testify/assert"
)

func setupElasticsearchEnv(t *testing.T, es *esboot.Elasticsearch) (env *goboot.AppEnv) {
	t.Helper()

	env = goboot.NewAppEnv("./testdata", "valid")
	assert.Nil(t, es.Configure(env))
	_ = es.IndexDelete(context.Background(), "test")
	_ = es.IndexDelete(context.Background(), es.MigrationsIndex)

	return env
}

func TestElasticsearch_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &esboot.Elasticsearch{}
	env := setupElasticsearchEnv(t, s)
	assert.Nil(t, s.Configure(env))
	assert.Nil(t, s.Init())
	assert.NotNil(t, s.Client)
	assert.NotNil(t, s.Config)
}

func TestElasticsearch_SuccessEnvs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &esboot.Elasticsearch{}
	_ = os.Setenv("ELASTICSEARCH_USERNAME", "elastic")
	_ = os.Setenv("ELASTICSEARCH_PASSWORD", "secret")

	err := s.Configure(goboot.NewAppEnv("./testdata", "using-env-vars"))
	assert.Nil(t, err)
	os.Clearenv()
}

func TestElasticsearch_ErrorNoAddresses(t *testing.T) {
	s := &esboot.Elasticsearch{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "no-addresses"))
	assert.EqualError(t, err, "config \"elasticsearch.addresses\" is required")
}

func TestElasticsearch_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &esboot.Elasticsearch{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "invalid-password"))
	assert.Contains(t, err.Error(), "expected 200 OK but got \"401 Unauthorized\" while retrieving elasticsearch info")
}
