package postgres_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/postgres"
	"github.com/stretchr/testify/assert"
)

func TestService_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &postgres.Service{}
	assert.Nil(t, s.Configure(goboot.NewAppContext("./testdata", "valid")))
	assert.Nil(t, s.Init())
	assert.Nil(t, s.Close())
}

func TestService_ErrorMissingConfig(t *testing.T) {
	s := &postgres.Service{}
	err := s.Configure(goboot.NewAppContext("./testdata", ""))
	assert.EqualError(t, err, "missing postgres configuration")
}

func TestService_ErrorMissingDSN(t *testing.T) {
	s := &postgres.Service{}
	err := s.Configure(goboot.NewAppContext("./testdata", "no-dsn"))
	assert.EqualError(t, err, "config \"postgres.dsn\" is required")
}

func TestService_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &postgres.Service{}
	err := s.Configure(goboot.NewAppContext("./testdata", "invalid-dsn"))
	assert.EqualError(t, err,
		"failed to connect to postgres \"postgres://postgres:REDACTED@1.2.3.4:5431/utils?sslmode=disable\" "+
			"after 5 retries: dial tcp 1.2.3.4:5431: i/o timeout",
	)
}
