package goboot

// AppService instantiates a singleton application service that is created
// on application boot and shutdown gracefully on application termination.
type AppService interface {
	// Configure is run when creating a new app context.
	Configure(ctx *AppContext) error

	// Init is run after all services have been configured. Use this to run
	// setup that is dependent on other services.
	//
	// The app will only start after all initializations are finished.
	Init() error

	// Close is run right before shutdown. The app waits until close resolves.
	// Any error is logged.
	Close() error

	Name() string
}
