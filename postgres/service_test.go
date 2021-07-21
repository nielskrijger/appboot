package postgres_test

import (
	"testing"

	"github.com/nielskrijger/goboot/context"
	"github.com/nielskrijger/goboot/postgres"
	"github.com/stretchr/testify/assert"
)

func TestService_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &postgres.Service{}
	s.Configure(context.NewAppContext("../test/conf", "postgres"))
	s.Init()
	s.Close()
}

func TestService_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	assert.Panics(t, func() {
		s := &postgres.Service{}
		s.Configure(context.NewAppContext("../test/conf", "postgres-invalid"))
		s.Init()
	})
}
