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

// Package lokitest provides test methods using which the plugins can validate whether they implement loki correctly or not.
package lokitest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/narahari92/loki/pkg/loki"
)

// Plugin represents a plugin implementation consisting of System, Destroyer and Killer.
type Plugin struct {
	// System implementation of plugin.
	System loki.System
	// Destroyer implementation of plugin.
	Destroyer loki.Destroyer
	// Killer implementation of plugin.
	Killer loki.Killer
}

// Configuration contains data needed to run tests through which a plugin can be validated.
type Configuration struct {
	// Identifiers is the list of resource IDs which should be killed to validate that system shows validation as false.
	Identifiers loki.Identifiers
	// DestroySection is sample destroy section to validate that destroyer can parse without error.
	DestroySection map[string]interface{}
}

// ValidateAll validates all functionalities of the plugin.
func ValidateAll(ctx context.Context, t *testing.T, plugin *Plugin, config *Configuration) {
	ValidateDestroyerParse(ctx, t, plugin, config)
	ValidateAfterSystemLoad(ctx, t, plugin)
	ValidateAfterKill(ctx, t, plugin, config)
}

// ValidateDestroyerParse validates that destroyer can parse destroy section without error.
func ValidateDestroyerParse(ctx context.Context, t *testing.T, plugin *Plugin, config *Configuration) {
	require.NotNil(t, config)
	require.NotNil(t, config.DestroySection)
	require.NotEqual(t, 0, len(config.DestroySection))

	identifiers, err := plugin.Destroyer.ParseDestroySection(config.DestroySection)
	require.NoError(t, err)
	require.NotNil(t, identifiers)
	require.NotEqual(t, 0, len(identifiers))
}

// ValidateAfterSystemLoad validates that validation passes if called immediately after system is loaded. So we check that
// when system is in desired state, validation succeeds.
func ValidateAfterSystemLoad(ctx context.Context, t *testing.T, plugin *Plugin) {
	err := plugin.System.Load(ctx)
	require.NoError(t, err)

	ok, err := plugin.System.Validate(ctx)
	require.NoError(t, err)
	require.Equal(t, true, ok)
}

// ValidateAfterKill validates that validation fails if called immediately after few resources are killed. So we check that
// when system is not in desired state, validation fails.
func ValidateAfterKill(ctx context.Context, t *testing.T, plugin *Plugin, config *Configuration) {
	require.NotNil(t, config)
	require.NotNil(t, config.Identifiers)
	require.NotEqual(t, 0, len(config.Identifiers))

	err := plugin.System.Load(ctx)
	require.NoError(t, err)

	err = plugin.Killer.Kill(ctx, config.Identifiers...)
	require.NoError(t, err)

	ok, err := plugin.System.Validate(ctx)
	if err == nil && ok {
		require.FailNow(t, "validate shouldn't succeed after kill")
	}
}
