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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAfter(t *testing.T) {
	readyFunc := After(2 * time.Second)
	now := time.Now()
	ready, err := readyFunc(context.Background())
	timeDiff := time.Now().UnixNano() - now.UnixNano()

	require.NoError(t, err)
	require.Equal(t, true, ready)
	require.Equal(t, int64(2), timeDiff/(1000*1000*1000))
}

func TestAllReady(t *testing.T) {
	readyFunc1 := After(2 * time.Second)
	readyFunc2 := After(4 * time.Second)
	allReady := AllReady(readyFunc1, readyFunc2)
	now := time.Now()
	ready, err := allReady(context.Background())
	timeDiff := time.Now().UnixNano() - now.UnixNano()

	require.NoError(t, err)
	require.Equal(t, true, ready)
	require.Equal(t, int64(6), timeDiff/(1000*1000*1000))
}
