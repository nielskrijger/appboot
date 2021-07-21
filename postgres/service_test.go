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
	assert.Nil(t, s.Configure(context.NewAppContext("../testdata/conf", "postgres")))
	assert.Nil(t, s.Init())
	assert.Nil(t, s.Close())
}

func TestService_ErrorOnMisconfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &postgres.Service{}
	err := s.Configure(context.NewAppContext("../testdata/conf", "empty"))
	assert.EqualError(t, err, "missing postgres configuration")
}

func TestService_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &postgres.Service{}
	err := s.Configure(context.NewAppContext("../testdata/conf", "postgres-invalid"))
	assert.EqualError(t, err, "failed to connect to postgres \"postgres://postgres:REDACTED@1.2.3.4:5431/utils?sslmode=disable\" after 5 retries: dial tcp 1.2.3.4:5431: i/o timeout")
}
