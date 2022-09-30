package acceptance

import (
	"os"
	"testing"
)

// package global test entry point (the Main() that runs only once and runs all the tests in this package)
//
// Note how we only start up the service once, then run all the acceptance tests. This is much faster
// than doing this every time.

func TestMain(m *testing.M) {
	err := tstSetup(validConfigurationPath)
	if err != nil {
		println("error during global acceptance test initialization - BAILING OUT!")
		os.Exit(1)
	}
	defer tstShutdown()

	code := m.Run()
	os.Exit(code)
}
