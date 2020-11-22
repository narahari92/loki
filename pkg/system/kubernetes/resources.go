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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func parseResources(resources []interface{}) ([]*ResourceIdentifier, error) {
	var identifiers []*ResourceIdentifier

	for _, resource := range resources {
		resIdent := &ResourceIdentifier{}
		k8sResource, ok := resource.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("resource defined should be a map type")
		}

		av, ok := k8sResource[apiVersionKey]
		if !ok {
			return nil, errors.Errorf(reqFieldErrMsg, apiVersionKey)
		}

		apiVersion, ok := av.(string)
		if !ok {
			return nil, errors.Errorf(strTypeErrMsg, apiVersionKey)
		}

		gv, err := schema.ParseGroupVersion(apiVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse apiVersion '%s'", apiVersion)
		}

		resIdent.Group = gv.Group
		resIdent.Version = gv.Version

		kindValue, ok := k8sResource[kindKey]
		if !ok {
			return nil, errors.Errorf(reqFieldErrMsg, kindKey)
		}

		kind, ok := kindValue.(string)
		if !ok {
			return nil, errors.Errorf(strTypeErrMsg, kindKey)
		}

		resIdent.Kind = kind

		if nameValue, ok := k8sResource[nameKey]; ok {
			name, ok := nameValue.(string)
			if !ok {
				return nil, errors.Errorf(strTypeErrMsg, nameKey)
			}

			resIdent.Name = name
		}

		if namespaceValue, ok := k8sResource[namespaceKey]; ok {
			namespace, ok := namespaceValue.(string)
			if !ok {
				return nil, errors.Errorf(strTypeErrMsg, namespaceKey)
			}

			resIdent.Namespace = namespace
		}

		identifiers = append(identifiers, resIdent)
	}

	return identifiers, nil
}
