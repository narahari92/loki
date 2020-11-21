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

import "context"

// Hook can be used to perform user defined actions at different stages of chaos execution such as log collection,
// fetching system state etc.
type Hook struct {
	preReady       func(context.Context) error
	postReady      func(context.Context) error
	preSystemLoad  func(context.Context) error
	postSystemLoad func(context.Context) error
	preChaos       func(context.Context) error
	postChaos      func(context.Context) error
}

// HookOption is functional option implementation to create Hook.
type HookOption func(h *Hook)

// WithPreReady functional option is used to hook user defined action to be executed before ready condition is executed.
func WithPreReady(preReady func(context.Context) error) HookOption {
	return func(h *Hook) {
		h.preReady = preReady
	}
}

// WithPostReady functional option is used to hook user defined action to be executed after ready condition is executed.
func WithPostReady(postReady func(context.Context) error) HookOption {
	return func(h *Hook) {
		h.postReady = postReady
	}
}

// WithPreSystemLoad functional option is used to hook user defined action to be executed before systems are loaded.
func WithPreSystemLoad(preSystemLoad func(context.Context) error) HookOption {
	return func(h *Hook) {
		h.preSystemLoad = preSystemLoad
	}
}

// WithPostSystemLoad functional option is used to hook user defined action to be executed after systems are loaded.
func WithPostSystemLoad(postSystemLoad func(context.Context) error) HookOption {
	return func(h *Hook) {
		h.postSystemLoad = postSystemLoad
	}
}

// WithPreChaos functional option is used to hook user defined action to be executed before first chaos scenario is executed.
func WithPreChaos(preChaos func(context.Context) error) HookOption {
	return func(h *Hook) {
		h.preChaos = preChaos
	}
}

// WithPostChaos functional option is used to hook user defined action to be executed after last chaos scenario is executed.
func WithPostChaos(postChaos func(context.Context) error) HookOption {
	return func(h *Hook) {
		h.postChaos = postChaos
	}
}
