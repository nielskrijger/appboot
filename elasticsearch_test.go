package goboot_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/stretchr/testify/assert"
)

func TestElasticSearch_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.ElasticSearch{}
	assert.Nil(t, s.Configure(goboot.NewAppContext("./testdata/elasticsearch", "valid")))
	assert.Nil(t, s.Init())
	assert.NotNil(t, s.Client)
	assert.NotNil(t, s.Config)
}

func TestElasticSearch_ErrorMissingConfig(t *testing.T) {
	s := &goboot.ElasticSearch{}
	err := s.Configure(goboot.NewAppContext("./testdata/elasticsearch", ""))
	assert.EqualError(t, err, "missing elasticsearch configuration")
}

func TestElasticSearch_ErrorNoAddresses(t *testing.T) {
	s := &goboot.ElasticSearch{}
	err := s.Configure(goboot.NewAppContext("./testdata/elasticsearch", "no-addresses"))
	assert.EqualError(t, err, "config \"elasticsearch.addresses\" is required")
}

func TestElasticSearch_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.ElasticSearch{}
	err := s.Configure(goboot.NewAppContext("./testdata/elasticsearch", "invalid-password"))
	assert.Contains(t, err.Error(), "expected 200 OK but got \"401 Unauthorized\" while retrieving elasticsearch info")
}
