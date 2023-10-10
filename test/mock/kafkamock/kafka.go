package kafkamock

import (
	"context"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
)
import _ "github.com/go-git/go-git/v5"

type Impl struct {
	Callback repository.ReceiverCallback

	Recording []repository.UpdateEvent
}

func New() repository.Kafka {
	return &Impl{
		Callback:  func(_ repository.UpdateEvent) {},
		Recording: make([]repository.UpdateEvent, 0),
	}
}

func (r *Impl) IsKafka() bool {
	return true
}

func (r *Impl) Setup() error {
	return nil
}

func (r *Impl) Teardown() {
}

func (r *Impl) SubscribeIncoming(_ context.Context, callback repository.ReceiverCallback) error {
	r.Callback = callback
	return nil
}

func (r *Impl) Send(_ context.Context, event repository.UpdateEvent) error {
	r.Recording = append(r.Recording, event)
	return nil
}

func (r *Impl) StartReceiveLoop(ctx context.Context) error {
	return nil
}

// --- test helpers ---

func (r *Impl) Receive(incomingEvent repository.UpdateEvent) {
	r.Callback(incomingEvent)
}

func (r *Impl) Reset() {
	r.Recording = make([]repository.UpdateEvent, 0)
}
