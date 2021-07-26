package goboot_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/stretchr/testify/assert"
)

func TestRedis_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Redis{}
	assert.Nil(t, s.Configure(goboot.NewAppContext("./testdata/redis", "valid")))
	assert.Nil(t, s.Init())
	assert.Equal(t, "Redis<0.0.0.0:6379 db:3>", s.Client.String())
}

func TestRedis_ErrorMissingConfig(t *testing.T) {
	s := &goboot.Redis{}
	err := s.Configure(goboot.NewAppContext("./testdata/redis", ""))
	assert.EqualError(t, err, "missing redis configuration")
}

func TestRedis_ErrorEmptyURL(t *testing.T) {
	s := &goboot.Redis{}
	err := s.Configure(goboot.NewAppContext("./testdata/redis", "no-url"))
	assert.EqualError(t, err, "config \"redis.url\" is required")
}

func TestRedis_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Redis{}
	err := s.Configure(goboot.NewAppContext("./testdata/redis", "invalid"))
	assert.EqualError(t, err, "failed to connect to redis after 5 retries: dial tcp 1.2.3.4:6379: i/o timeout")
}
