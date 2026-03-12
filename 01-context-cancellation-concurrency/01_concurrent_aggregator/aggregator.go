package _1_concurrent_aggregator

import (
	"context"
	"fmt"
	"go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/order"
	"go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/profile"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

type FuncOption func(agg *UserAggregator)

type UserAggregator struct {
	profileService profile.Service
	orderService   order.Service

	timeout time.Duration
	logger  *slog.Logger
}

func New(profile profile.Service, order order.Service, opts ...FuncOption) *UserAggregator {
	agg := &UserAggregator{
		profileService: profile,
		orderService:   order,
		logger:         slog.Default(),
	}

	for _, opt := range opts {
		opt(agg)
	}
	return agg
}

func WithTimeout(timeout time.Duration) FuncOption {
	return func(agg *UserAggregator) {
		agg.timeout = timeout
	}
}

func WithLogger(logger *slog.Logger) FuncOption {
	return func(agg *UserAggregator) {
		agg.logger = logger
	}
}

type AggregatedResult struct {
	userName  string
	orderCost float64
}

func (agg *UserAggregator) Aggregate(ctx context.Context, userId int) ([]*AggregatedResult, error) {
	var (
		localCtx context.Context
		cancel   context.CancelFunc
		orders   []*order.Order
		profile  *profile.Profile
		res      []*AggregatedResult
	)

	if agg.timeout > 0 {
		localCtx, cancel = context.WithTimeout(ctx, agg.timeout)
	} else {
		localCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	agg.logger.Info("start aggregation", "user_id", userId)

	g, localCtx := errgroup.WithContext(localCtx)
	g.Go(func() error {
		var err error
		orders, err = agg.orderService.GetAll(localCtx, userId)
		if err != nil {
			agg.logger.Warn("failed to get all order", "user_id", userId, "err", err)
			return fmt.Errorf("failed to fetch orders: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		profile, err = agg.profileService.Get(localCtx, userId)
		if err != nil {
			agg.logger.Warn("failed to get profile", "user_id", userId, "err", err)
			return fmt.Errorf("failed to fetch profile: %w", err)
		}
		return nil
	})

	err := g.Wait()
	if err != nil {
		agg.logger.Error("aggregation exited with error", "user_id", userId, "err", err)
		return nil, err
	}

	for _, o := range orders {
		if o.UserId == userId {
			res = append(res, &AggregatedResult{
				profile.Name,
				o.Cost,
			})
		}
	}

	agg.logger.Info("aggregation completed", "user_id", userId)
	return res, nil
}
