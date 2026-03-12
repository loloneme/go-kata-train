package _1_concurrent_aggregator

import (
	"context"
	"fmt"
	ordermock "go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/order/mocks"
	profilemock "go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/profile/mocks"
	"time"
)

func main() {
	profile := &profilemock.MockProfileService{
		Delay: 1 * time.Second,
	}

	orders := &ordermock.MockOrderService{
		Delay: 1 * time.Second,
	}

	agg := New(profile, orders, WithTimeout(3*time.Second))
	res, err := agg.Aggregate(context.Background(), 1)

	fmt.Println(res, err)
}
