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
	"context"

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/narahari92/loki/pkg/loki"
)

// Killer provides functionality to delete kubernetes resources.
type Killer struct {
	// System is the kubernetes system on which the killer acts on.
	*System
}

// Kill deletes the kubernetes resources represented by identifiers.
func (k *Killer) Kill(ctx context.Context, identifiers ...loki.Identifier) error {
	for _, identifier := range identifiers {
		resourceIdentifier, ok := identifier.(*ResourceIdentifier)
		if !ok {
			return errors.New("unsupported identifier passed to kubernetes killer")
		}

		object := &unstructured.Unstructured{}
		object.SetGroupVersionKind(resourceIdentifier.GroupVersionKind)
		object.SetNamespace(resourceIdentifier.Namespace)
		object.SetName(resourceIdentifier.Name)

		if err := k.k8sClient.Delete(ctx, object); err != nil {
			if k8serrors.IsNotFound(err) || meta.IsNoMatchError(err) {
				continue
			}

			return errors.Wrapf(err, "failed to delete kubernetes resource")
		}
	}

	return nil
}
