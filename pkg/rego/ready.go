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
	"encoding/json"

	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"

	"github.com/narahari92/loki/pkg/loki"
)

// ReadyCond performs readiness check based on rego policy.
type ReadyCond struct {
	query  string
	policy string
	system loki.System
}

// Ready checks whether system is ready.
func (r *ReadyCond) Ready(ctx context.Context) (bool, error) {
	systemJSON, err := r.system.AsJSON(ctx, true)
	if err != nil {
		return false, errors.Wrap(err, "failed to get system state")
	}

	var input interface{}
	if err := json.Unmarshal(systemJSON, &input); err != nil {
		return false, errors.Wrap(err, "failed to unmarshal system json")
	}

	regoDefinition := rego.New(
		rego.Query(r.query),
		rego.Module("", r.policy),
	)

	query, err := regoDefinition.PrepareForEval(ctx)
	if err != nil {
		return false, errors.Wrap(err, "failed to prepare rego definition for evaluation")
	}

	rs, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, errors.Wrap(err, "failed to evaluate repo input")
	}

	if len(rs) != 1 || rs[0].Bindings == nil {
		return false, errors.New("rego evaluation returned no result for ready check")
	}

	ready := false
	ok := false

	for _, value := range rs[0].Bindings {
		ready, ok = value.(bool)
		if !ok {
			return false, errors.New("rego evaluation returned non boolean type for ready check")
		}
	}

	return ready, nil
}
