package redisboot_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/redisboot"
	"github.com/stretchr/testify/assert"
)

func TestRedis_Success(t *testing.T) {
	s := &redisboot.Redis{}
	assert.Nil(t, s.Configure(goboot.NewAppEnv("./testdata", "valid")))
	assert.Nil(t, s.Init())
	assert.Equal(t, "Redis<0.0.0.0:6379 db:3>", s.Client.String())
}

func TestRedis_ErrorMissingConfig(t *testing.T) {
	s := &redisboot.Redis{}
	err := s.Configure(goboot.NewAppEnv("./testdata", ""))
	assert.EqualError(t, err, "missing Redis configuration")
}

func TestRedis_ErrorEmptyURL(t *testing.T) {
	s := &redisboot.Redis{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "no-url"))
	assert.EqualError(t, err, "config \"redis.url\" is required")
}

func TestRedis_ErrorOnConnect(t *testing.T) {
	s := &redisboot.Redis{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "invalid"))
	assert.EqualError(t, err, "failed to connect to redis after 5 retries: dial tcp 1.2.3.4:6379: i/o timeout")
}
