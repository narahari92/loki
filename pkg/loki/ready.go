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

const (
	afterKey    = "after"
	allReadyKey = "allReady"
)

// After is a very simple ReadyFunc which sleeps for given duration and then marks systems as ready.
func After(duration time.Duration) ReadyFunc {
	return func(context.Context) (bool, error) {
		time.Sleep(duration)

		return true, nil
	}
}

// AfterParser returns ReadyParser with can parse After ReadyFunc
func AfterParser(config *Config) ReadyParser {
	var readyParser ReadyParserFunc

	readyParser = func(readyConf map[string]interface{}) (ReadyCond, error) {
		after := readyConf[afterKey]

		duration, err := parseDuration(afterKey, after)
		if err != nil {
			return nil, err
		}

		return After(duration), nil
	}

	return readyParser
}

// AllReady is a ReadyFunc which takes in multiple ReadyCond instances and marks systems as ready only
// when all ReadyCond are in ready state.
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
