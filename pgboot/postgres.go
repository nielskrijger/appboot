package pgboot

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nielskrijger/goboot"
	"github.com/rs/zerolog"
)

const (
	defaultPostgresConnectMaxRetries    = 5
	defaultPostgresConnectRetryDuration = 5 * time.Second
)

var (
	errMissingConfig = errors.New("missing Postgres configuration")
	errMissingDSN    = errors.New("config \"postgres.dsn\" is required")
)

type PostgresConfig struct {
	// DSN contains the PGX data source name, e.;g. postgres://user:password@host:port/dbname?query
	// see also https://github.com/golang-migrate/migrate/tree/master/database/pgx
	DSN string `yaml:"dsn"`

	// Number of retries upon initial connect. Default is 5 times. Set -1 to disable
	ConnectMaxRetries int `yaml:"connectMaxRetries"`

	// Time between retries for initial connect attempts. Default is 5 seconds.
	ConnectRetryDuration time.Duration `yaml:"connectRetryDuration"`
}

// Postgres implements the AppService interface.
type Postgres struct {
	MigrationsDir string // relative path to migrations directory, leave empty when no migrations

	DB *sqlx.DB

	config  *PostgresConfig
	log     zerolog.Logger
	confDir string
}

func (s *Postgres) Name() string {
	return "Postgres"
}

// Configure connects to postgres.
func (s *Postgres) Configure(env *goboot.AppEnv) error {
	s.log = env.Log
	s.confDir = env.ConfDir

	// unmarshal config and set defaults
	s.config = &PostgresConfig{}

	if !env.Config.InConfig("postgres") {
		return errMissingConfig
	}

	if !env.Config.IsSet("postgres.dsn") {
		return errMissingDSN
	}

	if err := env.Config.Sub("postgres").Unmarshal(s.config); err != nil {
		return fmt.Errorf("parsing Postgres configuration: %w", err)
	}

	if s.config.ConnectMaxRetries == 0 {
		s.config.ConnectMaxRetries = defaultPostgresConnectMaxRetries
	}

	if s.config.ConnectRetryDuration == 0*time.Second {
		s.config.ConnectRetryDuration = defaultPostgresConnectRetryDuration
	}

	// Setup DB connection pool
	if err := s.connect(); err != nil {
		return err
	}

	return nil
}

func (s *Postgres) connect() error {
	db, err := sqlx.Open("pgx", s.config.DSN)
	if err != nil {
		return fmt.Errorf("connection to postgres: %w", err)
	}

	s.DB = db

	// Check if we can connect to PostgreSQL
	return s.testConnectivity()
}

func (s *Postgres) testConnectivity() error {
	// parse url for logging purposes
	logURL, err := url.Parse(s.config.DSN)
	if err != nil {
		return fmt.Errorf("invalid Postgres dsn: %w", err)
	}

	logURL.User = url.UserPassword(logURL.User.Username(), "REDACTED")
	s.log.Info().Msgf("connecting to %s", logURL.String())

	for retries := 1; ; retries++ {
		// test connection
		if err := s.DB.Ping(); err != nil {
			if retries < s.config.ConnectMaxRetries {
				s.log.
					Warn().
					Err(err).
					Str("url", logURL.String()).
					Msgf("failed to connect to Postgres, retrying in %s", s.config.ConnectRetryDuration)
			} else {
				return fmt.Errorf(
					"failed to connect to Postgres %q after %d retries: %w",
					logURL.String(),
					s.config.ConnectMaxRetries,
					err,
				)
			}

			time.Sleep(s.config.ConnectRetryDuration)
		} else {
			s.log.Info().Msg("successfully connected to Postgres")

			break
		}
	}

	return nil
}

func (s *Postgres) Init() error {
	u, err := url.Parse(s.config.DSN)
	if err != nil {
		return fmt.Errorf("invalid postgres dsn: %w", err)
	}

	if s.MigrationsDir == "" {
		s.log.Info().Msg("skipping db migrations; no migrations directory set")
	} else if err := s.Migrate(u.String(), s.MigrationsDir); err != nil {
		return fmt.Errorf("running Postgres migrations: %w", err)
	}

	return nil
}

func (s *Postgres) Close() error {
	if err := s.DB.Close(); err != nil {
		return fmt.Errorf("closing %s service: %w", s.Name(), err)
	}

	return nil
}
