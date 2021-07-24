package redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/nielskrijger/goboot"
	"github.com/rs/zerolog"
)

const (
	defaultConnectMaxRetries    = 5
	defaultConnectRetryDuration = 5 * time.Second
)

var (
	errMissingConfig = errors.New("missing redis configuration")
	errMissingURL    = errors.New("config \"redis.url\" is required")
)

type Config struct {
	// Url contains hostname:port, e.g. localhost:6379
	URL string `yaml:"url"`

	// Password if left empty uses no empty
	Password string `yaml:"password"`

	// DB defaults to db 0
	DB int `yaml:"db"`

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int `yaml:"poolSize"`

	// Dial timeout for establishing new connections. Default is 5 seconds.
	DialTimeout time.Duration `yaml:"dialTimeout"`

	// Number of retries upon initial connect. Default is 5 times. Set -1 to disable
	ConnectMaxRetries int `yaml:"connectMaxRetries"`

	// Time between retries for initial connect attempts. Default is 5 seconds.
	ConnectRetryDuration time.Duration `yaml:"connectRetryDuration"`
}

// Service implements the AppService interface.
type Service struct {
	Client *redis.Client

	log zerolog.Logger
}

func (s *Service) Name() string {
	return "redis"
}

func (s *Service) Configure(ctx *goboot.AppContext) error {
	s.log = ctx.Log
	redisCfg := &Config{}

	if !ctx.Config.InConfig("redis") {
		return errMissingConfig
	}

	if !ctx.Config.IsSet("redis.url") {
		return errMissingURL
	}

	if err := ctx.Config.Sub("redis").Unmarshal(redisCfg); err != nil {
		return fmt.Errorf("parsing redis configuration: %w", err)
	}

	s.log.Info().Msgf("connecting to redis %q, db %d", redisCfg.URL, redisCfg.DB)

	opts := &redis.Options{
		Addr:     redisCfg.URL,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	}
	if redisCfg.DialTimeout != 0 {
		opts.DialTimeout = redisCfg.DialTimeout
	}

	if redisCfg.PoolSize != 0 {
		opts.PoolSize = redisCfg.PoolSize
	}

	s.Client = redis.NewClient(opts)

	if redisCfg.ConnectMaxRetries == 0 {
		redisCfg.ConnectMaxRetries = defaultConnectMaxRetries
	}

	if redisCfg.ConnectRetryDuration == 0*time.Second {
		redisCfg.ConnectRetryDuration = defaultConnectRetryDuration
	}

	return s.testConnectivity(redisCfg)
}

func (s *Service) testConnectivity(cfg *Config) error {
	for retries := 1; ; retries++ {
		if err := s.Client.Ping().Err(); err != nil {
			if retries < cfg.ConnectMaxRetries {
				s.log.Warn().
					Err(err).
					Str("url", cfg.URL).
					Int("db", cfg.DB).
					Msgf("failed to connect to redis, retrying in %s", cfg.ConnectRetryDuration)
			} else {
				return fmt.Errorf(
					"failed to connect to redis after %d retries: %w",
					cfg.ConnectMaxRetries,
					err,
				)
			}

			time.Sleep(cfg.ConnectRetryDuration)
		} else {
			s.log.Info().Msg("successfully connected to redis")

			break
		}
	}

	return nil
}

// Init implements AppService interface.
func (s *Service) Init() error {
	return nil
}

// Close is run right before shutdown. The app waits until close resolves.
func (s *Service) Close() error {
	if err := s.Client.Close(); err != nil {
		return fmt.Errorf("closing %s service: %w", s.Name(), err)
	}

	return nil
}
