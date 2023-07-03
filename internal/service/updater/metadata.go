package updater

import (
	"context"
	"errors"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
)

func (s *Impl) updateMetadata(ctx context.Context) ([]repository.UpdateEvent, error) {
	s.Logging.Logger().Ctx(ctx).Info().Print("refreshing metadata")

	events, err := s.Mapper.RefreshMetadata(ctx)
	if err != nil {
		s.totalErrorCounter.Inc()
		s.metadataErrorCounter.Inc()
		return events, err
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			s.Logging.Logger().Ctx(ctx).Warn().Print("timeout while updating metadata")
			s.totalErrorCounter.Inc()
			s.metadataErrorCounter.Inc()
			return events, err
		}
	}

	return events, nil
}
