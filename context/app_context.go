package context

import (
	"os"

	"github.com/nielskrijger/go-utils/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// AppContext contains all application-scoped variables
type AppContext struct {
	Config  *viper.Viper
	Log     zerolog.Logger
	ConfDir string

	services []AppService
}

// AppService instantiates a singleton application service that is created
// on application boot and shutdown gracefully on application termination.
type AppService interface {
	// Configure is run immediately when creating a new app context. Any error
	// during configuration should cause a panic.
	Configure(ctx *AppContext)

	// Init is run after all services have been configured. Use this to run
	// setup that is dependent on other services.
	//
	// The app will only start after all initializations are finished. Any
	// error should cause a panic.
	Init()

	// Close is run right before shutdown. The app waits until close resolves.
	// Any error should be logged and handled by the service itself.
	Close()
}

// NewAppContext creates an AppContext by loading configuration settings and
// setting up common connections to databases and queues.
func NewAppContext(confDir string, env string) *AppContext {
	logger := newLogger()
	log.Info().Str("env", env).Msgf("starting server")
	cfg := config.MustLoadConfig(logger, confDir, env)

	return &AppContext{
		ConfDir:  confDir,
		Config:   cfg,
		Log:      logger,
		services: make([]AppService, 0),
	}
}

func (ctx *AppContext) AddService(service AppService) {
	ctx.services = append(ctx.services, service)
}

// newLogger returns a new zerolog logger.
//
// By default returns a production logger, to log on DEBUG level set en var LOG_DEBUG=true.
func newLogger() zerolog.Logger {
	// use env var instead of config because no config is available at startup
	debug, ok := os.LookupEnv("LOG_DEBUG")
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if ok && (debug == "true" || debug == "1") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	return zerolog.New(os.Stdout)
}

// Configure sets up service settings.
func (ctx *AppContext) Configure() {
	ctx.Log.Info().Msg("starting app services configuration")
	for _, service := range ctx.services {
		service.Configure(ctx)
	}
	ctx.Log.Info().Msg("finished app services configuration")
}

// Init runs all app service initialization.
func (ctx *AppContext) Init() {
	ctx.Log.Info().Msg("starting app services initialization")
	for _, service := range ctx.services {
		service.Init()
	}
	ctx.Log.Info().Msg("finished app services initialization")
}

// Close cleans up any resources held by any app services.
func (ctx *AppContext) Close() {
	ctx.Log.Info().Msg("start closing app services")
	for _, service := range ctx.services {
		service.Close()
	}
	ctx.Log.Info().Msg("finished closing app services")
}
