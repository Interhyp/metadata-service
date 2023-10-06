package service

// Trigger triggers update runs in Updater.
//
// Trigger events occur on initial app startup (before it becomes healthy), and periodically
type Trigger interface {
	IsTrigger() bool
	Setup() error
	Teardown()
}
