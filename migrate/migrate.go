package migrate

import (
	"database/sql"
	"errors"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4" //nolint
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Load file-loader for migration files
	_ "github.com/lib/pq"                                // Required dependency for postgres driver
	"github.com/rs/zerolog"
)

type Printer interface {
	Printf(format string, v ...interface{})
}

// DatabaseLogger implements the Logger interface of golang-migrate.
type logger struct {
	logger zerolog.Logger
}

// Printf is like fmt.Printf
func (log *logger) Printf(format string, v ...interface{}) {
	log.logger.Info().Msgf(format, v...)
}

// Verbose should return true when verbose logging output is wanted
func (log *logger) Verbose() bool {
	return true
}

// MustMigrate runs Postgres migration files from specified migrations directory.
//
// Panics if anything went wrong.
func MustMigrate(log zerolog.Logger, dsn string, migrationsDir string) {
	contextLogger := logger{logger: log}

	dir, err := filepath.Abs(migrationsDir)
	if err != nil {
		panic(err)
	}

	contextLogger.Printf("running database migrations from %s", dir)

	// connect to postgres
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	// run migrations
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+dir,
		"postgres",
		driver,
	)
	if err != nil {
		panic(err)
	}

	m.Log = &contextLogger

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			contextLogger.Printf("database is up-to-date")
		} else {
			panic(err)
		}
	} else {
		contextLogger.Printf("completed database migrations")
	}
}
