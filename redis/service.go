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

var errMissingConfig = errors.New("missing redis configuration")

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

func (s *Service) Configure(ctx *goboot.AppContext) error { // nolint:funlen
	s.log = ctx.Log
	redisConf := &Config{}

	if !ctx.Config.InConfig("redis") {
		return errMissingConfig
	}

	if err := ctx.Config.Sub("redis").Unmarshal(redisConf); err != nil {
		return fmt.Errorf("parsing redis configuration: %w", err)
	}

	s.log.Info().Msgf("connecting to redis %q, db %d", redisConf.URL, redisConf.DB)

	opts := &redis.Options{
		Addr:     redisConf.URL,
		Password: redisConf.Password,
		DB:       redisConf.DB,
	}
	if redisConf.DialTimeout != 0 {
		opts.DialTimeout = redisConf.DialTimeout
	}

	if redisConf.PoolSize != 0 {
		opts.PoolSize = redisConf.PoolSize
	}

	s.Client = redis.NewClient(opts)

	if redisConf.ConnectMaxRetries == 0 {
		redisConf.ConnectMaxRetries = defaultConnectMaxRetries
	}

	if redisConf.ConnectRetryDuration == 0*time.Second {
		redisConf.ConnectRetryDuration = defaultConnectRetryDuration
	}

	for retries := 1; ; retries++ {
		if err := s.Client.Ping().Err(); err != nil {
			if retries < redisConf.ConnectMaxRetries {
				s.log.Warn().
					Err(err).
					Str("url", redisConf.URL).
					Int("db", redisConf.DB).
					Msgf("failed to connect to redis, retrying in %s", redisConf.ConnectRetryDuration)
			} else {
				return fmt.Errorf(
					"failed to connect to redis after %d retries: %w",
					redisConf.ConnectMaxRetries,
					err,
				)
			}

			time.Sleep(redisConf.ConnectRetryDuration)
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
