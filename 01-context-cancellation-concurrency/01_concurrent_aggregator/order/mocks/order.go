package mocks

import (
	"context"
	"fmt"
	"go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/order"
	"time"
)

type MockOrderService struct {
	Delay  time.Duration
	Fail   bool
	Orders []*order.Order
}

func (m *MockOrderService) GetAll(ctx context.Context, userId int) ([]*order.Order, error) {
	fmt.Println("Simulating orders search..")

	var res []*order.Order

	select {
	case <-time.After(m.Delay):
		if m.Fail {
			return res, fmt.Errorf("order service error")
		}
	case <-ctx.Done():
		return res, ctx.Err()
	}

	for _, o := range m.Orders {
		if o.UserId == userId {
			res = append(res, o)
		}
	}

	return res, nil
}
