/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wait

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestExecuteWithBackoff(t *testing.T) {
	tests := []struct {
		description string
		backoff     Backoff
		function    func(ctx context.Context) (bool, error)
		timeout     time.Duration
		err         error
		pass        bool
		minDuration time.Duration
		maxDuration time.Duration
	}{
		{
			description: "successful function execution with exponential backoff",
			backoff: &ExponentialBackoff{
				Duration: 1 * time.Second,
				Cap:      15 * time.Second,
				Factor:   2,
				Jitter:   0.3,
			},
			function: func(ctx context.Context) (bool, error) {
				return true, nil
			},
			timeout:     30 * time.Second,
			pass:        true,
			minDuration: 0 * time.Second,
			maxDuration: 1 * time.Second,
		},
		{
			description: "failed function execution with exponential backoff",
			backoff: &ExponentialBackoff{
				Duration: 1 * time.Second,
				Cap:      15 * time.Second,
				Factor:   2,
				Jitter:   0.3,
			},
			function: func(ctx context.Context) (bool, error) {
				return false, errors.New("failed execution")
			},
			timeout:     20 * time.Second,
			err:         errors.New("failed execution"),
			pass:        false,
			minDuration: 20 * time.Second,
			maxDuration: 35 * time.Second,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			before := time.Now()
			ok, err := ExecuteWithBackoff(context.Background(), test.backoff, test.function, test.timeout)
			after := time.Now()

			if test.err != nil {
				require.Error(t, err)
				require.Equal(t, test.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.pass, ok)
			require.LessOrEqual(t, int64(after.Sub(before)), int64(test.maxDuration))
			require.GreaterOrEqual(t, int64(after.Sub(before)), int64(test.minDuration))
		})
	}
}
