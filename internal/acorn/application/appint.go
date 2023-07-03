package application

// ApplicationName is used as a default for logging very early errors when the configuration isn't read yet.
const ApplicationName = "metadata"

// Application is the central singleton representing the entire application.
type Application interface {
	IsApplication() bool

	// Register registers all Acorns that make up the application.
	//
	// after a call to Register you can override the Acorn constructors in the registry, e.g. for mocking.
	//
	// if not already called, Run will call this for you.
	Register()

	// Create instantiates all Acorns, but does not connect them.
	//
	// after a call to Create you can replace Acorns by name in the registry, e.g. for testing/mocking.
	//
	// if not already called, Run will call this for you.
	Create()

	// Assemble wires up all Acorns.
	//
	// to avoid a circular dependency with logging, this also parses the configuration, but does not validate it.
	//
	// if not already called, Run will call this for you.
	Assemble() error

	// Run runs the application, including setup and teardown phase
	//
	// returns the exit code - we do not call os.Exit inside
	Run() int
}
