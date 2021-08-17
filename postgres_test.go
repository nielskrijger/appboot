package goboot_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/stretchr/testify/assert"
)

func TestPostgres_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Postgres{}
	assert.Nil(t, s.Configure(goboot.NewAppEnv("./testdata/postgres", "valid")))
	assert.Nil(t, s.Init())
	assert.Nil(t, s.Close())
}

func TestPostgres_ErrorMissingConfig(t *testing.T) {
	s := &goboot.Postgres{}
	err := s.Configure(goboot.NewAppEnv("./testdata/postgres", ""))
	assert.EqualError(t, err, "missing postgres configuration")
}

func TestPostgres_ErrorMissingDSN(t *testing.T) {
	s := &goboot.Postgres{}
	err := s.Configure(goboot.NewAppEnv("./testdata/postgres", "no-dsn"))
	assert.EqualError(t, err, "config \"postgres.dsn\" is required")
}

func TestPostgres_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Postgres{}
	err := s.Configure(goboot.NewAppEnv("./testdata/postgres", "invalid-dsn"))
	assert.EqualError(t, err,
		"failed to connect to postgres \"postgres://postgres:REDACTED@1.2.3.4:5431/utils?sslmode=disable\" "+
			"after 5 retries: dial tcp 1.2.3.4:5431: i/o timeout",
	)
}
