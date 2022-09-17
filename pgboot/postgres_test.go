package pgboot_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/pgboot"
	"github.com/stretchr/testify/assert"
)

func TestPostgres_Success(t *testing.T) {
	s := &pgboot.Postgres{}
	assert.Nil(t, s.Configure(goboot.NewAppEnv("./testdata", "valid")))
	assert.Nil(t, s.Init())
	assert.Nil(t, s.Close())
}

func TestPostgres_ErrorMissingConfig(t *testing.T) {
	s := &pgboot.Postgres{}
	err := s.Configure(goboot.NewAppEnv("./testdata", ""))
	assert.EqualError(t, err, "missing Postgres configuration")
}

func TestPostgres_ErrorMissingDSN(t *testing.T) {
	s := &pgboot.Postgres{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "no-dsn"))
	assert.EqualError(t, err, "config \"postgres.dsn\" is required")
}

func TestPostgres_ErrorOnConnect(t *testing.T) {
	s := &pgboot.Postgres{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "invalid-dsn"))
	assert.EqualError(t, err,
		"failed to connect to Postgres \"postgres://postgres:REDACTED@1.2.3.4:5431/utils?sslmode=disable\" "+
			"after 5 retries: dial tcp 1.2.3.4:5431: i/o timeout",
	)
}
