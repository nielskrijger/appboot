package pgboot

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4" //nolint
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rs/zerolog"

	// Load file-loader for migration files.
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Required dependency for postgres driver.
	_ "github.com/lib/pq"
)

type PostgresMigratePrinter interface {
	Printf(format string, v ...any)
}

// DatabaseLogger implements the Logger interface of golang-migrate.
type logger struct {
	logger zerolog.Logger
}

// Printf is like fmt.Printf.
func (log *logger) Printf(format string, v ...any) {
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
	log := logger{logger: s.log}

	dir, err := filepath.Abs(migrations)
	if err != nil {
		return fmt.Errorf("reading migrations path: %w", err)
	}

	log.Printf("running Postgres migrations from %s", dir)

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
		return fmt.Errorf("open Postgres connection for golang-migrate: %w", err)
	}

	// setup migrations connection
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+dir,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("connecting to Postgres for migrations: %w", err)
	}

	m.Log = &log

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("Postgres database is up-to-date")
		} else {
			return fmt.Errorf("running Postgres migrations: %w", err)
		}
	} else {
		log.Printf("completed Postgres migrations")
	}

	return nil
}
