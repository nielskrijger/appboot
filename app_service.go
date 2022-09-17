package goboot

// AppService instantiates a singleton application service that is created
// on application boot and shutdown gracefully on application termination.
type AppService interface {
	// Configure is run when a new app context is created. Use this to load
	// configuration settings.
	//
	// Any error will cause a panic. The app waits for all services to be
	// configured before calling Init.
	Configure(env *AppEnv) error

	// Init is run after all services have been configured. Use this to run
	// setup that is dependent on other services.
	//
	// Any error will cause a panic. The app starts after all initializations
	// are finished.
	Init() error

	// Close is run right before shutdown. The app waits until close resolves.
	//
	// Any returned error is logged but will not cause a panic or exit.
	Close() error

	// Name returns the name of the service used for logging purposes.
	Name() string
}
