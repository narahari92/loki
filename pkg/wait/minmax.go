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

// MinMaxBackoff implements Backoff and at each step provides duration between Min and Max values.
type MinMaxBackoff struct {
	// Min value of the duration generated at each step.
	Min time.Duration
	// Max value of the duration generated at each step.
	Max time.Duration
}

// Step returns the next duration at which to execute functionality.
func (m *MinMaxBackoff) Step() time.Duration {
	randInt := rand.Int63n(int64(m.Max) - int64(m.Min))

	return time.Duration(int64(m.Min) + randInt)
}
