package notification

import (
	"context"
	"fmt"
	"time"
)

type Processor struct {
	Repo     *Repository
	Delivery Delivery
	Now      func() time.Time
}

type NoopDelivery struct{}

func (NoopDelivery) Send(context.Context, DeliverySubscription, Notice) error { return nil }

func (p Processor) RunOnce(ctx context.Context) (int, error) {
	if p.Repo == nil {
		return 0, fmt.Errorf("notification repository unavailable")
	}
	now := time.Now()
	if p.Now != nil {
		now = p.Now()
	}
	if _, err := p.Repo.CreateDueNotices(ctx, now); err != nil {
		return 0, err
	}
	if p.Delivery == nil {
		return 0, nil
	}
	subs, err := p.Repo.DuePushSubscriptions(ctx, now, 50)
	if err != nil {
		return 0, err
	}
	delivered := 0
	for _, sub := range subs {
		items, _, err := p.Repo.List(ctx, sub.UserID, 1, time.Time{})
		if err != nil {
			return delivered, err
		}
		if len(items) == 0 {
			continue
		}
		if err := p.Delivery.Send(ctx, sub, items[0]); err != nil {
			if updateErr := p.Repo.RecordDeliveryFailure(ctx, sub.ID, IsExpiredDeliveryError(err), now); updateErr != nil {
				return delivered, updateErr
			}
			continue
		}
		if err := p.Repo.RecordDeliverySuccess(ctx, sub.ID, now); err != nil {
			return delivered, err
		}
		delivered++
	}
	return delivered, nil
}
