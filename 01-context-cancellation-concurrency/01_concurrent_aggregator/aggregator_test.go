package _1_concurrent_aggregator

import (
	"bytes"
	"context"
	"go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/order"
	orderMock "go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/order/mocks"
	"go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/profile"
	profileMock "go-train/01-context-cancellation-concurrency/01_concurrent_aggregator/profile/mocks"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregator(t *testing.T) {
	type Input struct {
		profileService        *profileMock.MockProfileService
		orderService          *orderMock.MockOrderService
		userId                int
		timeout               time.Duration
		parentContextDuration time.Duration
	}

	type Expected struct {
		aggregatedProfiles []*AggregatedResult
		err                bool
	}

	type TestCase struct {
		name     string
		input    Input
		expected Expected
	}

	basicProfiles := []*profile.Profile{
		{Id: 1, Name: "Alice"},
		{Id: 2, Name: "Bob"},
		{Id: 3, Name: "Charlie"},
		{Id: 4, Name: "Dave"},
		{Id: 5, Name: "Eva"},
	}

	var emptyAggregateProfiles []*AggregatedResult
	testCases := []TestCase{
		{
			name: "no errors, profile service in time, order service in time, more than one order",
			input: Input{
				profileService: &profileMock.MockProfileService{
					10 * time.Millisecond,
					false,
					basicProfiles,
				},
				orderService: &orderMock.MockOrderService{
					20 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 1, 100.0},
						{2, 1, 20.6},
						{3, 3, 30.79},
					},
				},
				userId:  1,
				timeout: 100 * time.Millisecond,
			},
			expected: Expected{
				[]*AggregatedResult{{"Alice", 100.0}, {"Alice", 20.6}},
				false,
			},
		},
		{
			"no errors, profile service in time, order service in time, one order",
			Input{
				&profileMock.MockProfileService{
					10 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					20 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 3, 30.79},
					},
				},
				1,
				100 * time.Millisecond,
				0,
			},
			Expected{
				[]*AggregatedResult{{"Alice", 100.0}},
				false,
			},
		},
		{
			"no errors, profile service in time, order service in time, no orders",
			Input{
				&profileMock.MockProfileService{
					10 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					20 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 2, 100.0},
						{3, 3, 30.79},
					},
				},
				1,
				100 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				false,
			},
		},
		{
			"no errors, profile service in time, order service timeout, more than one order",
			Input{
				&profileMock.MockProfileService{
					10 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					120 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 1, 30.79},
					},
				},
				1,
				100 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				true,
			},
		},
		{
			"no errors, profile service in time, order service timeout, one order",
			Input{
				&profileMock.MockProfileService{
					10 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					120 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 2, 30.79},
					},
				},
				1,
				100 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				true,
			},
		},
		{
			"no errors, profile service in time, order service timeout, no orders",
			Input{
				&profileMock.MockProfileService{
					10 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					120 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 3, 100.0},
						{3, 2, 30.79},
					},
				},
				1,
				100 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				true,
			},
		},
		{
			"no errors, profile service timeout, order service in time, more than one order",
			Input{
				&profileMock.MockProfileService{
					120 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					20 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 1, 30.79},
					},
				},
				1,
				100 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				true,
			},
		},
		{
			"no service timeout, no errors, more than one order",
			Input{
				&profileMock.MockProfileService{
					120 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					150 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 1, 30.79},
					},
				},
				1,
				0 * time.Millisecond,
				0,
			},
			Expected{
				[]*AggregatedResult{{"Alice", 100.0}, {"Alice", 30.79}},
				false,
			},
		},
		{
			"no service timeout, profile error, more than one order",
			Input{
				&profileMock.MockProfileService{
					120 * time.Millisecond,
					true,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					150 * time.Millisecond,
					false,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 1, 30.79},
					},
				},
				1,
				0 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				true,
			},
		},
		{
			"no service timeout, order error, more than one order",
			Input{
				&profileMock.MockProfileService{
					70 * time.Millisecond,
					false,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					100 * time.Millisecond,
					true,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 1, 30.79},
					},
				},
				1,
				0 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				true,
			},
		},
		{
			"no service timeout, profile error, order error, profile and order take same time",
			Input{
				&profileMock.MockProfileService{
					120 * time.Millisecond,
					true,
					basicProfiles,
				},
				&orderMock.MockOrderService{
					120 * time.Millisecond,
					true,
					[]*order.Order{
						{1, 1, 100.0},
						{3, 1, 30.79},
					},
				},
				1,
				0 * time.Millisecond,
				0,
			},
			Expected{
				emptyAggregateProfiles,
				true,
			},
		},
		{
			"order service error propagates correctly",
			Input{
				&profileMock.MockProfileService{10 * time.Millisecond, false, basicProfiles},
				&orderMock.MockOrderService{10 * time.Millisecond, true, nil}, // Error here
				1,
				100 * time.Millisecond,
				0,
			},
			Expected{
				nil,
				true,
			},
		},
		{
			"timeout error propagates correctly",
			Input{
				&profileMock.MockProfileService{200 * time.Millisecond, false, basicProfiles},
				&orderMock.MockOrderService{10 * time.Millisecond, false, nil},
				1,
				50 * time.Millisecond,
				0,
			},
			Expected{
				nil,
				true,
			},
		},
		{
			"parent context cancellation propagates correctly",
			Input{
				&profileMock.MockProfileService{200 * time.Millisecond, false, basicProfiles},
				&orderMock.MockOrderService{10 * time.Millisecond, false, nil},
				1,
				500 * time.Millisecond,
				100 * time.Millisecond,
			},
			Expected{
				nil,
				true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			if tc.input.parentContextDuration > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.input.parentContextDuration)
				defer cancel()
			}

			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			u := New(tc.input.profileService, tc.input.orderService, WithTimeout(tc.input.timeout), WithLogger(logger))

			res, err := u.Aggregate(ctx, tc.input.userId)
			logOutput := buf.String()
			if tc.expected.err {
				require.Error(t, err)
				assert.Contains(t, logOutput, "error")
			} else {
				require.NoError(t, err)
				assert.Contains(t, logOutput, "user_id")
			}

			assert.Equal(t, tc.expected.aggregatedProfiles, res)
		})
	}

}
