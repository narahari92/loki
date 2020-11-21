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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScenario(t *testing.T) {
	RegisterTestSystem(t)

	conf, err := ioutil.ReadFile("../../testdata/sample-config.yaml")
	require.NoError(t, err)

	configuration := NewConfig()

	err = configuration.Parse([]byte(conf))
	require.NoError(t, err)

	scenarios := make([]Identifiers, 0)

	for systemName, provider := range configuration.scenarioProviders {
		require.Equal(t, "testing", systemName)
		system := configuration.systems[systemName]

		for {
			scenario, ok := provider.scenario(system)
			if !ok {
				break
			}

			scenarios = append(scenarios, scenario.identifiers)
		}
	}

	require.Equal(t, 4, len(scenarios))
	require.Equal(t, Identifiers{TestIdentifier("resource2")}, scenarios[0])
	require.Equal(t, Identifiers{TestIdentifier("resource2"), TestIdentifier("resource4")}, scenarios[1])

	require.NotEqual(t, Identifiers{TestIdentifier("resource2")}, scenarios[2])
	require.NotEqual(t, Identifiers{TestIdentifier("resource2"), TestIdentifier("resource4")}, scenarios[2])
	require.NotEqual(t, Identifiers{TestIdentifier("resource4"), TestIdentifier("resource2")}, scenarios[2])
	require.NotEqual(t, Identifiers{TestIdentifier("resource1")}, scenarios[2])
	require.NotEqual(t, Identifiers{TestIdentifier("resource2"), TestIdentifier("resource3")}, scenarios[2])
	require.NotEqual(t, Identifiers{TestIdentifier("resource3"), TestIdentifier("resource2")}, scenarios[2])

	require.NotEqual(t, Identifiers{TestIdentifier("resource2")}, scenarios[3])
	require.NotEqual(t, Identifiers{TestIdentifier("resource2"), TestIdentifier("resource4")}, scenarios[3])
	require.NotEqual(t, Identifiers{TestIdentifier("resource4"), TestIdentifier("resource2")}, scenarios[3])
	require.NotEqual(t, Identifiers{TestIdentifier("resource1")}, scenarios[3])
	require.NotEqual(t, Identifiers{TestIdentifier("resource2"), TestIdentifier("resource3")}, scenarios[3])
	require.NotEqual(t, Identifiers{TestIdentifier("resource3"), TestIdentifier("resource2")}, scenarios[2])
	require.NotEqual(t, scenarios[2], scenarios[3])
}
