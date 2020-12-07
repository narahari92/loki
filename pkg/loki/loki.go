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
	"strings"
	"sync"
)

var (
	availableSystems = make(map[string]func() System)
	systemsMx        sync.Mutex

	availableDestroyers = make(map[string]Destroyer)
	destroyersMx        sync.Mutex

	availableKillers = make(map[string]func(System) (Killer, error))
	killersMx        sync.Mutex

	readyParsers = make(map[string]func(*Config) ReadyParser)
	readyMx      sync.Mutex
)

// ID uniquely represents any resource or operation in a system.
type ID string

// System is the core interface which represents any execution environment such as Kubernetes, AWS etc.
// Plugin implementations need to implement this interface which handles things such as parsing system
// definition, loading desired state, performing validation etc.
type System interface {
	// Parse  parses the definition of system.
	Parse(map[string]interface{}) error
	// Load loads the state of system as per its definition.
	Load(context.Context) error
	// Validate validates at any point of time whether the actual state of system matches with its desired state
	// as determined by ReadyCond.
	Validate(context.Context) (bool, error)
	// Identifiers return Identifier values of all resources in the system.
	Identifiers() Identifiers
	// AsJSON returns the json representation of the state of the system. If `reload` is set to `true`, state of the system
	// will be reloaded before preparing json representation of system.
	AsJSON(ctx context.Context, reload bool) ([]byte, error)
}

// Destroyer parses the single section of destroy whether it be exclusions or scenarios. Plugin implementations
// need to implement this interface.
type Destroyer interface {
	// ParseDestroySection parses the any section under destroy block.
	ParseDestroySection(map[string]interface{}) (Identifiers, error)
}

// DestroyerFunc is the syntax sugar for single method Destroyer interface so that a simple function can implement
// Destroyer interface.
type DestroyerFunc func(map[string]interface{}) (Identifiers, error)

// ParseDestroySection calls d(m).
func (d DestroyerFunc) ParseDestroySection(m map[string]interface{}) (Identifiers, error) {
	return d(m)
}

// Killer kills the given identifiers. Definition of kill depends on system. For example, in kubernetes it could be
// deleting resource and for networks it could creating disconnection between systems.
type Killer interface {
	// Kill kills given identifiers.
	Kill(context.Context, ...Identifier) error
}

// KillerFunc is the syntax sugar for single method Killer interface so that a simple function can implement
// Killer interface.
type KillerFunc func(context.Context, ...Identifier) error

// KillerFunc calls k(ctx, i).
func (k KillerFunc) Kill(ctx context.Context, i ...Identifier) error {
	return k(ctx, i...)
}

// ReadyCond defines the condition where in all the systems are considered to be in desired state.
type ReadyCond interface {
	// Ready checks whether system has reached desired state.
	Ready(context.Context) (bool, error)
}

// ReadyFunc is the syntax sugar for  single method ReadyCond interface so that a simple function can implement
// ReadyCond interface.
type ReadyFunc func(context.Context) (bool, error)

// Ready calls r(ctx).
func (r ReadyFunc) Ready(ctx context.Context) (bool, error) {
	return r(ctx)
}

// ReadyParser parses the ready section to create ReadyCond.
type ReadyParser interface {
	// Parse parses the ready section and creates ReadyCond.
	Parse(map[string]interface{}) (ReadyCond, error)
}

// ReadyParserFunc is the syntax sugar for  single method ReadyParser interface so that a simple function can implement
// ReadyParser interface.
type ReadyParserFunc func(map[string]interface{}) (ReadyCond, error)

// Parse calls r(readyConf).
func (r ReadyParserFunc) Parse(readyConf map[string]interface{}) (ReadyCond, error) {
	return r(readyConf)
}

// Identifier uniquely  identifies any resource or operation in a particular system.
type Identifier interface {
	// ID returns the unique identifier of resource or operation in a particular system.
	ID() ID
}

// Identifiers is group of Identifier which can be used to create scenarios for creating chaos, excluding the chaos
// scenarios etc.
type Identifiers []Identifier

// String method returns string representation of Identifiers.
func (idents Identifiers) String() string {
	sb := strings.Builder{}
	sb.WriteString("[\n")

	for _, ident := range idents {
		sb.WriteString("{")
		sb.WriteString(string(ident.ID()))
		sb.WriteString("}\n")
	}

	sb.WriteString("]")

	return sb.String()
}

// RegisterSystem is used by plugins to register custom systems.
func RegisterSystem(name string, system func() System) {
	systemsMx.Lock()
	defer systemsMx.Unlock()

	availableSystems[name] = system
}

// RegisterDestroyer is used by plugins to register custom destroyers.
func RegisterDestroyer(name string, destroyer Destroyer) {
	destroyersMx.Lock()
	defer destroyersMx.Unlock()

	availableDestroyers[name] = destroyer
}

// RegisterKiller is used by plugins to register custom killers.
func RegisterKiller(name string, killer func(System) (Killer, error)) {
	killersMx.Lock()
	defer killersMx.Unlock()

	availableKillers[name] = killer
}

// RegisterReadyParser registers ReadyParser creating functions that can be used by config file.
func RegisterReadyParser(key string, parser func(*Config) ReadyParser) {
	readyMx.Lock()
	defer readyMx.Unlock()

	readyParsers[key] = parser
}
