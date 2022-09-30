package kafkamock

import (
	"context"
	"github.com/Interhyp/metadata-service/acorns/repository"
)
import _ "github.com/go-git/go-git/v5"

type Impl struct {
	Callback repository.ReceiverCallback

	Recording []repository.UpdateEvent
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
