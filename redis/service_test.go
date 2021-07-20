package redis

import (
	"testing"

	"github.com/nielskrijger/go-utils/context"
	"github.com/stretchr/testify/assert"
)

func TestService_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &Service{}
	s.Configure(context.NewAppContext("../test/conf", "redis"))
	s.Init()

	assert.Equal(t, "Redis<0.0.0.0:6379 db:3>", s.Client.String())
}

func TestService_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	assert.Panics(t, func() {
		s := &Service{}
		s.Configure(context.NewAppContext("../test/conf", "redis-invalid"))
	})
}
