package postgres

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-pg/pg"
	"github.com/nielskrijger/goboot"
	"github.com/rs/zerolog"
)

const (
	defaultConnectMaxRetries    = 5
	defaultConnectRetryDuration = 5 * time.Second
)

var errMissingConfig = errors.New("missing postgres configuration")

type Config struct {
	// DSN contains hostname:port, e.g. localhost:6379
	DSN string `yaml:"dsn"`

	// Number of seconds before first connect attempt times out.
	ConnectTimeout int `yaml:"connectTimeout"`

	// Number of retries upon initial connect. Default is 5 times. Set -1 to disable
	ConnectMaxRetries int `yaml:"connectMaxRetries"`

	// Time between retries for initial connect attempts. Default is 5 seconds.
	ConnectRetryDuration time.Duration `yaml:"connectRetryDuration"`
}

// Service implements the AppService interface.
type Service struct {
	DB     *pg.DB
	Config *Config

	log     zerolog.Logger
	confDir string
}

type dbLogger struct {
	log zerolog.Logger
}

func (d *dbLogger) BeforeQuery(q *pg.QueryEvent) {}

func (d *dbLogger) AfterQuery(q *pg.QueryEvent) {
	str, err := q.FormattedQuery()
	if err != nil {
		d.log.Error().Err(err).Msg("error retrieving query")
	} else {
		d.log.Debug().Msg(str)
	}
}

type healtcheckResult struct {
	Result int
}

func (s *Service) Name() string {
	return "postgres"
}

// Configure connects to postgres and logs connection info for
// debugging connectivity issues.
func (s *Service) Configure(ctx *goboot.AppContext) error {
	s.log = ctx.Log
	s.confDir = ctx.ConfDir

	// unmarshal config and set defaults
	s.Config = &Config{}

	if !ctx.Config.InConfig("postgres") {
		return errMissingConfig
	}

	if err := ctx.Config.Sub("postgres").UnmarshalExact(s.Config); err != nil {
		return fmt.Errorf("failed parsing postgres configuration: %w", err)
	}

	if s.Config.ConnectMaxRetries == 0 {
		s.Config.ConnectMaxRetries = defaultConnectMaxRetries
	}

	if s.Config.ConnectRetryDuration == 0*time.Second {
		s.Config.ConnectRetryDuration = defaultConnectRetryDuration
	}

	// log dsn for debugging purposes
	if err := s.connect(); err != nil {
		return err
	}

	// print SQL queries when debug logging is on
	if ctx.Log.Debug().Enabled() {
		s.DB.AddQueryHook(&dbLogger{log: s.log})
	}

	return nil
}

func (s *Service) connect() error {
	// parse url for logging purposes
	logURL, err := url.Parse(s.Config.DSN)
	if err != nil {
		return fmt.Errorf("invalid postgres dsn: %w", err)
	}

	logURL.User = url.UserPassword(logURL.User.Username(), "REDACTED")
	s.log.Info().Msgf("connecting to %s", logURL.String())

	// parse
	pgOptions, err := pg.ParseURL(s.Config.DSN)
	if err != nil {
		return fmt.Errorf("could not parse postgres DSN: %w", err)
	}

	pgOptions.DialTimeout = time.Duration(s.Config.ConnectTimeout) * time.Second

	for retries := 1; ; retries++ {
		s.DB = pg.Connect(pgOptions)

		// test connection
		if _, err := s.DB.Query(&healtcheckResult{}, "SELECT 1 AS result"); err != nil {
			if retries < s.Config.ConnectMaxRetries {
				s.log.
					Warn().
					Err(err).
					Str("url", logURL.String()).
					Msgf("failed to connect to postgres, retrying in %s", s.Config.ConnectRetryDuration)
			} else {
				return fmt.Errorf(
					"failed to connect to postgres %q after %d retries: %w",
					logURL.String(),
					s.Config.ConnectMaxRetries,
					err,
				)
			}

			time.Sleep(s.Config.ConnectRetryDuration)
		} else {
			s.log.Info().Msg("successfully connected to postgres")

			break
		}
	}

	return nil
}

func (s *Service) Init() error {
	u, err := url.Parse(s.Config.DSN)
	if err != nil {
		return fmt.Errorf("invalid postgres dsn: %w", err)
	}

	q := u.Query()
	q.Set("connect_timeout", strconv.Itoa(s.Config.ConnectTimeout))
	u.RawQuery = q.Encode()

	if err := Migrate(s.log, u.String(), s.confDir+"/migrations"); err != nil {
		return fmt.Errorf("running postgres migrations: %w", err)
	}

	return nil
}

func (s *Service) Close() error {
	if err := s.DB.Close(); err != nil {
		return fmt.Errorf("closing %s service: %w", s.Name(), err)
	}

	return nil
}
