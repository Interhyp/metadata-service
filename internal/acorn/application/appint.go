package application

// ApplicationName is used as a default for logging very early errors when the configuration isn't read yet.
const ApplicationName = "metadata"

// Application is the central singleton representing the entire application.
type Application interface {
	IsApplication() bool

	// Run runs the application, including setup and teardown phase
	//
	// returns the exit code - we do not call os.Exit inside
	Run() int
}
