package wait

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMinMaxBackoff(t *testing.T) {
	tests := []struct {
		description string
		backoff     *MinMaxBackoff
	}{
		{
			description: "different units for min and max",
			backoff: &MinMaxBackoff{
				Min: 500 * time.Millisecond,
				Max: 1 * time.Second,
			},
		},
		{
			description: "same units for min and max",
			backoff: &MinMaxBackoff{
				Min: 1 * time.Second,
				Max: 2 * time.Second,
			},
		},
		{
			description: "extremely small gap for min and max",
			backoff: &MinMaxBackoff{
				Min: 100 * time.Nanosecond,
				Max: 105 * time.Nanosecond,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			minMaxDuration := test.backoff.Step()
			require.GreaterOrEqual(t, int64(minMaxDuration), int64(test.backoff.Min))
			require.Less(t, int64(minMaxDuration), int64(test.backoff.Max))
		})
	}
}
