package mocks

import (
	"context"
	"fmt"
	"go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/profile"
	"time"
)

type MockProfileService struct {
	Delay    time.Duration
	Fail     bool
	Profiles []*profile.Profile
}

func (m *MockProfileService) Get(ctx context.Context, id int) (*profile.Profile, error) {
	fmt.Println("Simulating profile search..")

	select {
	case <-time.After(m.Delay):
		if m.Fail {
			return nil, fmt.Errorf("profile service error")
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	for _, p := range m.Profiles {
		if p.Id == id {
			return p, nil
		}
	}
	return nil, nil
}
