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

package rego

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/narahari92/loki/pkg/loki"
)

func TestReadyCond(t *testing.T) {
	tests := []struct {
		description string
		policy      string
		query       string
		ready       bool
		expectedErr error
	}{
		{
			description: "ready successfully evaluating to true",
			policy: `package policy
default ready = false

ready {
    resource1_ready
    resource2_ready
}

resource1_ready {
	"resource1" == input[_]
}

resource2_ready {
	"resource2" == input[_]
}
`,
			query: "ready := data.policy.ready",
			ready: true,
		},
		{
			description: "ready successfully evaluating to false",
			policy: `package policy
default ready = false

ready {
    resource1_ready
    resource7_ready
}

resource1_ready {
	"resource1" == input[_]
}

resource7_ready {
	"resource7" == input[_]
}
`,
			query: "ready := data.policy.ready",
			ready: false,
		},
		{
			description: "failed to evaluate because of empty query",
			policy: `package policy
default ready = false

ready {
    resource1_ready
    resource2_ready
}

resource1_ready {
	"resource1" == input[_]
}

resource2_ready {
	"resource2" == input[_]
}
`,
			expectedErr: errors.New("failed to prepare rego definition for evaluation: cannot evaluate empty query"),
		},
		{
			description: "failed to evaluate because of non boolean evaluation",
			policy: `package policy

ready[resource] {
    resource = input[1]
}
`,
			query:       "ready := data.policy.ready",
			expectedErr: errors.New("rego evaluation returned non boolean type for ready check"),
		},
	}

	system := &loki.TestSystem{
		Resources: map[loki.TestIdentifier]bool{
			"resource1": true,
			"resource2": true,
			"resource3": true,
			"resource4": true,
			"resource5": true,
			"resource6": true,
		},
		State: make(map[loki.TestIdentifier]bool),
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			readyCond := &ReadyCond{
				query:  tc.query,
				policy: tc.policy,
				system: system,
			}

			ready, err := readyCond.Ready(context.Background())
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.ready, ready)
		})

	}
}
