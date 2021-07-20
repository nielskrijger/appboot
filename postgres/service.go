package postgres

import (
	"github.com/rs/zerolog"
	"net/url"
	"strconv"
	"time"

	"github.com/go-pg/pg"
	"github.com/nielskrijger/goboot/context"
	"github.com/nielskrijger/goboot/migrate"
)

const (
	defaultConnectMaxRetries    = 5
	defaultConnectRetryDuration = 5 * time.Second
)

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

// configurePostgres connects to postgres and logs connection info for
// debugging connectivity issues.
func (s *Service) Configure(ctx *context.AppContext) {
	s.log = ctx.Log
	s.confDir = ctx.ConfDir

	// unmarshal config and set defaults
	s.Config = &Config{}
	if err := ctx.Config.Sub("postgres").Unmarshal(s.Config); err != nil {
		s.log.Panic().Err(err).Msg("failed parsing redis configuration")
	}
	if s.Config.ConnectMaxRetries == 0 {
		s.Config.ConnectMaxRetries = defaultConnectMaxRetries
	}
	if s.Config.ConnectRetryDuration == 0*time.Second {
		s.Config.ConnectRetryDuration = defaultConnectRetryDuration
	}

	// log dsn for debugging purposes
	s.connect()

	// print SQL queries when debug logging is on
	if ctx.Log.Debug().Enabled() {
		s.DB.AddQueryHook(&dbLogger{log: s.log})
	}
}

func (s *Service) connect() {
	// parse url for logging purposes
	logURL, err := url.Parse(s.Config.DSN)
	if err != nil {
		s.log.Panic().Err(err).Msg("invalid postgres dsn")
	}
	logURL.User = url.UserPassword(logURL.User.Username(), "REDACTED")
	s.log.Info().Msgf("connecting to %s", logURL.String())

	// parse
	pgOptions, err := pg.ParseURL(s.Config.DSN)
	if err != nil {
		s.log.Panic().Err(err).Msg("could not parse postgres DSN")
	}
	pgOptions.DialTimeout = time.Duration(s.Config.ConnectTimeout) * time.Second

	for retries := 1; ; retries++ {
		s.DB = pg.Connect(pgOptions)

		// test connection
		if _, err := s.DB.Query(&healtcheckResult{}, "SELECT 1 AS result"); err != nil {
			subLog := s.log.With().Err(err).Str("url", logURL.String()).Logger()
			if retries < s.Config.ConnectMaxRetries {
				subLog.Warn().Msgf("failed to connect to postgres, retrying in %s", s.Config.ConnectRetryDuration)
			} else {
				subLog.Panic().Msgf("failed to connect to postgres after %d retries", s.Config.ConnectMaxRetries)
			}
			time.Sleep(s.Config.ConnectRetryDuration)
		} else {
			s.log.Info().Msg("successfully connected to postgres")
			break
		}
	}
}

func (s *Service) Init() {
	u, err := url.Parse(s.Config.DSN)
	if err != nil {
		s.log.Panic().Err(err).Msg("invalid dsn")
	}
	q := u.Query()
	q.Set("connect_timeout", strconv.Itoa(s.Config.ConnectTimeout))
	u.RawQuery = q.Encode()
	migrate.MustMigrate(s.log, u.String(), s.confDir+"/migrations")
}

func (s *Service) Close() {
	if err := s.DB.Close(); err != nil {
		s.log.Error().Err(err).Msg("failed closing postgres connection gracefully")
	}
}
