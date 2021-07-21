package goboot

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// AppContext contains all application-scoped variables.
type AppContext struct {
	Config   *viper.Viper
	Log      zerolog.Logger
	ConfDir  string
	Services []AppService
}

// NewAppContext creates an AppContext by loading configuration settings and
// setting up common connections to databases and queues.
func NewAppContext(confDir string, env string) *AppContext {
	logger := newLogger()
	logger.Info().Str("env", env).Msgf("starting server")

	cfg := MustLoadConfig(logger, confDir, env)

	return &AppContext{
		ConfDir:  confDir,
		Config:   cfg,
		Log:      logger,
		Services: make([]AppService, 0),
	}
}

func (ctx *AppContext) AddService(service AppService) {
	ctx.Services = append(ctx.Services, service)
}

// newLogger returns a new zerolog logger.
//
// By default returns a production logger, to log on DEBUG level set en var LOG_DEBUG=true.
func newLogger() zerolog.Logger {
	// use env var instead of config because no config is available at startup
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	debug, ok := os.LookupEnv("LOG_DEBUG")

	if ok && (debug == "true" || debug == "1") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return zerolog.New(os.Stdout)
}

// Configure sets up service settings.
func (ctx *AppContext) Configure() {
	ctx.Log.Info().Msg("starting app services configuration")

	for _, service := range ctx.Services {
		if err := service.Configure(ctx); err != nil {
			ctx.Log.Panic().Err(err).Msgf("failed to configure service %s", service.Name())
		}
	}

	ctx.Log.Info().Msg("finished app services configuration")
}

// Init runs all app service initialization.
func (ctx *AppContext) Init() {
	ctx.Log.Info().Msg("starting app services initialization")

	for _, service := range ctx.Services {
		if err := service.Init(); err != nil {
			ctx.Log.Panic().Err(err).Msgf("failed to initialize service %s", service.Name())
		}
	}

	ctx.Log.Info().Msg("finished app services initialization")
}

// Close cleans up any resources held by any app services.
func (ctx *AppContext) Close() {
	ctx.Log.Info().Msg("start closing app services")

	for _, service := range ctx.Services {
		if err := service.Close(); err != nil {
			ctx.Log.Error().Err(err).Msgf("failed to gracefully close service %s", service.Name())
		}
	}

	ctx.Log.Info().Msg("finished closing app services")
}
