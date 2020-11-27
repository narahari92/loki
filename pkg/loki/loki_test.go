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
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type TestIdentifier string

func (t TestIdentifier) ID() ID {
	return ID(t)
}

type TestSystem struct {
	resources map[TestIdentifier]bool
	state     map[TestIdentifier]bool
}

func (t *TestSystem) Parse(m map[string]interface{}) (err error) {
	defer func() {
		if a := recover(); a != nil {
			err = errors.New("panic occurred while parsing")
		}
	}()

	resources := m["resources"].([]interface{})

	for _, resource := range resources {
		t.resources[TestIdentifier(resource.(string))] = true
	}

	return nil
}

func (t *TestSystem) Load(ctx context.Context) error {
	for id, val := range t.resources {
		t.state[id] = val
	}

	return nil
}

func (t *TestSystem) Validate(ctx context.Context) (bool, error) {
	if len(t.resources) != len(t.state) {
		return false, nil
	}

outerLoop:
	for resId, resVal := range t.resources {
		for stateId, stateVal := range t.state {
			if resId == stateId && resVal == stateVal {
				continue outerLoop
			}
		}

		return false, nil
	}
	return true, nil
}

func (t *TestSystem) Identifiers() Identifiers {
	var identifiers Identifiers

	for resource := range t.resources {
		identifiers = append(identifiers, resource)
	}

	return identifiers
}

func (t *TestSystem) AsJSON() ([]byte, error) {
	return json.Marshal(t.state)
}

func DestroyerTest() DestroyerFunc {
	return func(m map[string]interface{}) (identifiers Identifiers, err error) {
		defer func() {
			if a := recover(); a != nil {
				err = errors.New("panic occurred while parsing")
			}
		}()

		resources := m["resources"].([]interface{})

		for _, resource := range resources {
			identifiers = append(identifiers, TestIdentifier(resource.(string)))
		}

		return identifiers, nil
	}
}

type TestKiller struct {
	System *TestSystem
}

func (t *TestKiller) Kill(_ context.Context, identifiers ...Identifier) (err error) {
	defer func() {
		if a := recover(); a != nil {
			err = errors.New("panic occurred while parsing")
		}
	}()

	for _, identifier := range identifiers {
		if _, ok := t.System.state[identifier.(TestIdentifier)]; ok {
			delete(t.System.state, identifier.(TestIdentifier))
		}
	}

	return nil
}

func RegisterTestSystem(t *testing.T) {
	RegisterSystem("test-system", func() System {
		return &TestSystem{
			resources: make(map[TestIdentifier]bool),
			state:     make(map[TestIdentifier]bool),
		}
	})
	RegisterDestroyer("test-system", DestroyerTest())
	RegisterKiller("test-system", func(system System) (Killer, error) {
		testSystem, ok := system.(*TestSystem)
		require.Equal(t, true, ok)

		return &TestKiller{
			System: testSystem,
		}, nil
	})
}
