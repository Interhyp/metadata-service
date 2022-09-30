package kafkamock

import (
	"github.com/Interhyp/metadata-service/acorns/repository"
	"github.com/StephanHCB/go-autumn-acorn-registry/api"
)

// --- implementing Acorn ---

func New() auacornapi.Acorn {
	return &Impl{
		Callback:  func(_ repository.UpdateEvent) {},
		Recording: make([]repository.UpdateEvent, 0),
	}
}

func (r *Impl) IsKafka() bool {
	return true
}

func (r *Impl) AcornName() string {
	return repository.KafkaAcornName
}

func (r *Impl) AssembleAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}

func (r *Impl) SetupAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}

func (r *Impl) TeardownAcorn(_ auacornapi.AcornRegistry) error {
	return nil
}
