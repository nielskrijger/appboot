package goboot

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4" //nolint
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Load file-loader for migration files.
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Required dependency for postgres driver.
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

type PostgresMigratePrinter interface {
	Printf(format string, v ...interface{})
}

// DatabaseLogger implements the Logger interface of golang-migrate.
type logger struct {
	logger zerolog.Logger
}

// Printf is like fmt.Printf.
func (log *logger) Printf(format string, v ...interface{}) {
	log.logger.Info().Msgf(format, v...)
}

// Verbose should return true when verbose logging output is wanted.
func (log *logger) Verbose() bool {
	return true
}

// Migrate runs Postgres migration files from specified migrations directory.
//
// Panics if anything went wrong.
func (s *Postgres) Migrate(dsn string, migrations string) error {
	ctxLog := logger{logger: s.log}

	dir, err := filepath.Abs(migrations)
	if err != nil {
		return fmt.Errorf("reading migrations path: %w", err)
	}

	ctxLog.Printf("running database migrations from %s", dir)

	// connect to postgres
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}()

	p := &postgres.Postgres{}

	driver, err := p.Open(dsn)
	if err != nil {
		return fmt.Errorf("open postgres connection for golang-migrate: %w", err)
	}

	// setup migrations connection
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+dir,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("connecting to postgres for migrations: %w", err)
	}

	m.Log = &ctxLog

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			ctxLog.Printf("database is up-to-date")
		} else {
			return fmt.Errorf("running migrations: %w", err)
		}
	} else {
		ctxLog.Printf("completed database migrations")
	}

	return nil
}
