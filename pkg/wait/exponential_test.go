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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExponentialBackoff_Step(t *testing.T) {
	tests := []struct {
		description string
		exponential *ExponentialBackoff
		minDuration time.Duration
		maxDuration time.Duration
	}{
		{
			description: "exponential struct",
			exponential: &ExponentialBackoff{
				Duration: 1 * time.Minute,
				Cap:      10 * time.Minute,
				Factor:   2,
				Jitter:   0.3,
			},
			minDuration: 120 * time.Second,
			maxDuration: 156 * time.Second,
		},
		{
			description: "exponential struct beyond cap",
			exponential: &ExponentialBackoff{
				Duration: 6 * time.Minute,
				Cap:      10 * time.Minute,
				Factor:   2,
				Jitter:   0.3,
			},
			minDuration: 10 * time.Minute,
			maxDuration: 10 * time.Minute,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			duration := test.exponential.Step()
			require.LessOrEqual(t, int64(test.minDuration), int64(duration))
			require.GreaterOrEqual(t, int64(test.maxDuration), int64(duration))
		})
	}
}
