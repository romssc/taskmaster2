package update

import (
	"context"
	"taskmaster2/service2/internal/domain"
	"taskmaster2/service2/internal/pkg/broker/kafkaa"
)

func EventHandler() kafkaa.EventHandler {
	return func(ctx context.Context, event domain.Event) error {
		if err := u.Update(ctx, event); err != nil {
			return err
		}
		return nil
	}
}
