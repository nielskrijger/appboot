package esboot_test

import (
	"context"
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/esboot"
	"github.com/stretchr/testify/assert"
)

func setupElasticsearchEnv(t *testing.T, es *esboot.Elasticsearch) *goboot.AppEnv {
	t.Helper()

	env := goboot.NewAppEnv("./testdata", "valid")
	assert.Nil(t, es.Configure(env))
	_ = es.IndexDelete(context.Background(), "test")
	_ = es.IndexDelete(context.Background(), es.MigrationsIndex)

	return env
}

func TestElasticsearch_Success(t *testing.T) {
	s := &esboot.Elasticsearch{}
	env := setupElasticsearchEnv(t, s)
	assert.Nil(t, s.Configure(env))
	assert.Nil(t, s.Init())
	assert.NotNil(t, s.Client)
	assert.NotNil(t, s.Config)
}

func TestElasticsearch_SuccessEnvs(t *testing.T) {
	s := &esboot.Elasticsearch{}

	t.Setenv("ELASTICSEARCH_USERNAME", "elastic")
	t.Setenv("ELASTICSEARCH_PASSWORD", "secret")

	err := s.Configure(goboot.NewAppEnv("./testdata", "using-env-vars"))
	assert.Nil(t, err)
}

func TestElasticsearch_ErrorNoAddresses(t *testing.T) {
	s := &esboot.Elasticsearch{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "no-addresses"))
	assert.EqualError(t, err, "config \"elasticsearch.addresses\" is required")
}

func TestElasticsearch_ErrorOnConnect(t *testing.T) {
	s := &esboot.Elasticsearch{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "invalid-password"))
	assert.Contains(t, err.Error(), "expected 200 OK but got \"401 Unauthorized\" while retrieving Elasticsearch info")
}
