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

package kubernetes

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/narahari92/loki/pkg/loki"
)

const (
	kubernetesResource = "loki:kubernetes-resource"
	system             = "kubernetes"
)

// ResourceIdentifier implements loki.Identifier for kubernetes resources.
type ResourceIdentifier struct {
	// GroupVersionKind identifies the type of kubernetes resource.
	schema.GroupVersionKind
	// Name represents the name of kubernetes resource.
	Name string
	// Namespace represents the namespace in which kubernetes resources lies. It should be empty for cluster scoped resoruces.
	Namespace string
}

// ID returns the unique identifier of kubernetes resource.
func (r *ResourceIdentifier) ID() loki.ID {
	return loki.ID(kubernetesResource + ":" + r.GroupVersionKind.String() + ", " + r.Namespace + "/" + r.Name)
}

// Register registers the kubernetes system, destroyer and killer with loki.
func Register() {
	loki.RegisterSystem(system, func() loki.System {
		return NewSystem()
	})
	loki.RegisterDestroyer(system, Destroyer())
	loki.RegisterKiller(system, func(system loki.System) (loki.Killer, error) {
		kubernetesSystem, ok := system.(*System)
		if !ok {
			return nil, errors.New("unsupported system passed to instantiate kubernetes killer")
		}

		return &Killer{
			System: kubernetesSystem,
		}, nil
	})
}
