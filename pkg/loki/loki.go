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
)

type ID string

type System interface {
	Parse(map[string]interface{}) error
	Load(context.Context) error
	Validate(context.Context) (bool, error)
	Identifiers() Identifiers
}

type Destroyer interface {
	ParseDestroySection(map[string]interface{}) (Identifiers, error)
}

type DestroyerFunc func(map[string]interface{}) (Identifiers, error)

func (d DestroyerFunc) ParseDestroySection(m map[string]interface{}) (Identifiers, error) {
	return d(m)
}

type Killer interface {
	Kill(context.Context, ...Identifier) error
}

type KillerFunc func(context.Context, ...Identifier) error

func (k KillerFunc) Kill(ctx context.Context, i ...Identifier) error {
	return k(ctx, i...)
}

type ReadyCond interface {
	Ready(context.Context) (bool, error)
}

type ReadyFunc func(context.Context) (bool, error)

func (r ReadyFunc) Ready(ctx context.Context) (bool, error) {
	return r(ctx)
}

type Identifier interface {
	ID() ID
}

type Identifiers []Identifier

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

func RegisterSystem(name string, system func() System) {
	systemsMx.Lock()
	defer systemsMx.Unlock()

	availableSystems[name] = system
}

func RegisterDestroyer(name string, destroyer Destroyer) {
	destroyersMx.Lock()
	defer destroyersMx.Unlock()

	availableDestroyers[name] = destroyer
}

func RegisterKiller(name string, killer func(System) (Killer, error)) {
	killersMx.Lock()
	defer killersMx.Unlock()

	availableKillers[name] = killer
}
