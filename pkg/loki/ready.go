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

package loki

import (
	"context"
	"time"
)

func After(duration time.Duration) ReadyFunc {
	return func(context.Context) (bool, error) {
		time.Sleep(duration)

		return true, nil
	}
}

func AllReady(readyConditions ...ReadyCond) ReadyFunc {
	return func(ctx context.Context) (bool, error) {
		for _, readyCond := range readyConditions {
			ready, err := readyCond.Ready(ctx)
			if err != nil || !ready {
				return ready, err
			}
		}

		return true, nil
	}
}
