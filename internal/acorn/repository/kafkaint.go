package repository

import "context"

// Kafka is the central singleton representing the kafka messaging bus.
type Kafka interface {
	IsKafka() bool

	// Setup only connects the producer, the consumer is connected with StartReceiveLoop.
	Setup() error
	// Teardown will close both producer and consumer if they have been connected.
	Teardown()

	// SubscribeIncoming allows you to register a callback that is called whenever a message is received from the Kafka bus.
	//
	// Note, we currently only allow a single callback, so calling this multiple times will overwrite the callback.
	// Use this during application setup.
	SubscribeIncoming(ctx context.Context, callback ReceiverCallback) error

	// Send sends an UpdateEvent that originates in this application to the Kafka bus.
	Send(ctx context.Context, event UpdateEvent) error

	// StartReceiveLoop starts a background goroutine that calls the subscribed callback when messages come in
	StartReceiveLoop(ctx context.Context) error
}

type ReceiverCallback func(event UpdateEvent)

type UpdateEvent struct {
	Affected EventAffects `json:"affected"`

	// ISO-8601 UTC date time at which this information was committed.
	TimeStamp string `json:"timeStamp"`
	// The git commit hash this information was committed under.
	CommitHash string `json:"commitHash"`
}

type EventAffects struct {
	OwnerAliases   []string `json:"ownerAliases"`
	ServiceNames   []string `json:"serviceNames"`
	RepositoryKeys []string `json:"repositoryKeys"`
}
