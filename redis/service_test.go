package redis_test

import (
	"testing"

	"github.com/nielskrijger/goboot/context"
	"github.com/nielskrijger/goboot/redis"
	"github.com/stretchr/testify/assert"
)

func TestService_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &redis.Service{}
	s.Configure(context.NewAppContext("../testdata/conf", "redis"))
	s.Init()

	assert.Equal(t, "Redis<0.0.0.0:6379 db:3>", s.Client.String())
}

func TestService_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	assert.Panics(t, func() {
		s := &redis.Service{}
		s.Configure(context.NewAppContext("../testdata/conf", "redis-invalid"))
	})
}
