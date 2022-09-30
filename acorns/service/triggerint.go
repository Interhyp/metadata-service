package service

const TriggerAcornName = "trigger"

// Trigger triggers update runs in Updater.
//
// Trigger events occur on initial app startup (before it becomes healthy), and periodically
type Trigger interface {
	IsTrigger() bool
}
