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
	"time"
)

// Backoff interface specifies the operation which provides next duration at which to execute a functionality.
type Backoff interface {
	// Step returns the next duration at which to execute functionality.
	Step() time.Duration
}

// ExecuteWithBackoff executes given function with backoff in case of failure.
func ExecuteWithBackoff(ctx context.Context, backoff Backoff, run func(ctx context.Context) (bool, error), timeout time.Duration) (bool, error) {
	start := time.Now()

	for {
		ok, err := run(ctx)

		if timeout > 0 && time.Now().After(start.Add(timeout)) {
			return ok, err
		}

		if err == nil && ok {
			return ok, err
		}

		sleepDuration := backoff.Step()

		time.Sleep(sleepDuration)
	}
}
