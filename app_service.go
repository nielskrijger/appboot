package goboot

// AppService instantiates a singleton application service that is created
// on application boot and shutdown gracefully on application termination.
type AppService interface {
	// Configure is run when a new app context is created. Use this to load
	// configuration settings and do initial setup.
	//
	// Any error will cause a panic. The app waits for all services to be
	// Configure'd before calling Init.
	Configure(ctx *AppEnv) error

	// Init is run after all services have been configured. Use this to run
	// setup that is dependent on other services.
	//
	// If the service does not depend on any other services you should do
	// setup during Configure instead of Init.
	//
	// Any error will cause a panic. The app starts after all initializations
	// are finished.
	Init() error

	// Close is run right before shutdown. The app waits until close resolves.
	//
	// Any returned error is logged but will not cause a panic or exit.
	Close() error

	Name() string
}
