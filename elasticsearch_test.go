package goboot_test

import (
	"context"
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/stretchr/testify/assert"
)

func setupElasticsearchEnv(t *testing.T, es *goboot.Elasticsearch) (env *goboot.AppEnv) {
	t.Helper()

	env = goboot.NewAppEnv("./testdata/elasticsearch", "valid")
	assert.Nil(t, es.Configure(env))
	_ = es.IndexDelete(context.Background(), "test")
	_ = es.IndexDelete(context.Background(), es.MigrationsIndex)

	return env
}

func TestElasticsearch_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Elasticsearch{}
	env := setupElasticsearchEnv(t, s)
	assert.Nil(t, s.Configure(env))
	assert.Nil(t, s.Init())
	assert.NotNil(t, s.Client)
	assert.NotNil(t, s.Config)
}

func TestElasticsearch_ErrorMissingConfig(t *testing.T) {
	s := &goboot.Elasticsearch{}
	err := s.Configure(goboot.NewAppEnv("./testdata/elasticsearch", ""))
	assert.EqualError(t, err, "missing \"elasticsearch\" configuration")
}

func TestElasticsearch_ErrorNoAddresses(t *testing.T) {
	s := &goboot.Elasticsearch{}
	err := s.Configure(goboot.NewAppEnv("./testdata/elasticsearch", "no-addresses"))
	assert.EqualError(t, err, "config \"elasticsearch.addresses\" is required")
}

func TestElasticsearch_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Elasticsearch{}
	err := s.Configure(goboot.NewAppEnv("./testdata/elasticsearch", "invalid-password"))
	assert.Contains(t, err.Error(), "expected 200 OK but got \"401 Unauthorized\" while retrieving elasticsearch info")
}
