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
	"math/rand"
	"time"
)

// ExponentialBackoff implements Backoff and provides exponentially increasing duration with each iteration.
type ExponentialBackoff struct {
	// Duration represents initial duration and subsequent durations in each step.
	Duration time.Duration
	// Cap is the maximum duration at which backoff is capped at.
	Cap time.Duration
	// Factor represents the multiplication factor at which backoff duration grows at each step.
	Factor float64
	// Jitter is maximum percentage variation in duration at which it grows at each step.
	Jitter float64
}

// Step returns the next duration at which to execute functionality.
func (e *ExponentialBackoff) Step() time.Duration {
	if e.Duration <= 0 {
		e.Duration = 10 * time.Second
	}

	duration := float64(e.Duration)
	if e.Factor > 0.0 {
		duration = float64(e.Duration) * e.Factor
	}

	if e.Jitter > 0.0 && e.Jitter <= 1.0 {
		duration += rand.Float64() * e.Jitter * duration
	}

	e.Duration = time.Duration(duration)

	if e.Cap > 0 && e.Cap < e.Duration {
		e.Duration = e.Cap
	}

	return e.Duration
}
