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
	"github.com/pkg/errors"

	"github.com/narahari92/loki/pkg/loki"
)

// Destroyer parses the destroy section i.e. exclusion and scenario for kubernetes system.
func Destroyer() loki.DestroyerFunc {
	return func(destroySection map[string]interface{}) (loki.Identifiers, error) {
		resources, ok := destroySection[resourcesKey]
		if !ok {
			return nil, errors.Errorf("'%s' field must be defined for kubernetes system", resourcesKey)
		}

		k8sResources, ok := resources.([]interface{})
		if !ok {
			return nil, errors.Errorf("'%s' field should be of type array", resourcesKey)
		}

		resourceIdentifiers, err := parseResources(k8sResources)
		if err != nil {
			return nil, err
		}

		var identifiers loki.Identifiers

		for _, resourceIdentifier := range resourceIdentifiers {
			identifiers = append(identifiers, resourceIdentifier)
		}

		return identifiers, nil
	}
}
