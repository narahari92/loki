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
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/narahari92/loki/pkg/loki"

	"github.com/narahari92/loki/pkg/lokitest"
)

func TestSystem(t *testing.T) {
	destroySectionYaml := `
system: test-system
resources:
- apiVersion: apps/v1
  kind: Deployment
  name: deploy1
  namespace: test-ns
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  name: test-cr2
`
	destroySection := make(map[string]interface{})

	err := yaml.Unmarshal([]byte(destroySectionYaml), &destroySection)
	require.NoError(t, err)

	k8sClient, identifiers := createClientAndIdentifiers(t)
	system := NewSystem()
	system.k8sClient = k8sClient
	system.resourceIdentifiers = identifiers
	killer := &Killer{
		System: system,
	}
	destroyer := Destroyer()
	k8sPlugin := &lokitest.Plugin{
		System:    system,
		Destroyer: destroyer,
		Killer:    killer,
	}
	configuration := &lokitest.Configuration{
		Identifiers: loki.Identifiers{
			&ResourceIdentifier{
				GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
				Name:             "deploy2",
				Namespace:        "test-ns",
			},
		},
		DestroySection: destroySection,
	}

	lokitest.ValidateAll(context.Background(), t, k8sPlugin, configuration)
}

func createClientAndIdentifiers(t *testing.T) (client.Client, []*ResourceIdentifier) {
	k8sClient, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	var objects []runtime.Object
	var identifiers []*ResourceIdentifier

	objects = append(objects, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns",
		},
	})
	identifiers = append(identifiers, &ResourceIdentifier{
		GroupVersionKind: schema.GroupVersionKind{Version: "v1", Kind: "Namespace"},
		Name:             "test-ns",
	})

	objects = append(objects, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy1",
			Namespace: "test-ns",
			Labels: map[string]string{
				"name": "deploy1",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "deploy1",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "deploy1",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "test-container1",
							Image: "test-image1",
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: v1.ConditionTrue,
				},
				{
					Type:   appsv1.DeploymentProgressing,
					Status: v1.ConditionTrue,
				},
			},
		},
	})
	identifiers = append(identifiers, &ResourceIdentifier{
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Name:             "deploy1",
		Namespace:        "test-ns",
	})

	objects = append(objects, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy2",
			Namespace: "test-ns",
			Labels: map[string]string{
				"name": "deploy2",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "deploy2",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "deploy2",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "test-container2",
							Image: "test-image2",
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: v1.ConditionFalse,
				},
				{
					Type:   appsv1.DeploymentProgressing,
					Status: v1.ConditionTrue,
				},
			},
		},
	})
	identifiers = append(identifiers, &ResourceIdentifier{
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Name:             "deploy2",
		Namespace:        "test-ns",
	})

	objects = append(objects, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test-cm1",
		},
		Data: map[string]string{
			"key1": "value1",
		},
	})
	identifiers = append(identifiers, &ResourceIdentifier{
		GroupVersionKind: schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
		Name:             "test-cm1",
		Namespace:        "test-ns",
	})

	objects = append(objects, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test-cm2",
		},
		Data: map[string]string{
			"key2": "value2",
		},
	})

	objects = append(objects, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cr1",
		},
	})
	identifiers = append(identifiers, &ResourceIdentifier{
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		Name:             "test-cr1",
	})

	objects = append(objects, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cr2",
		},
	})
	identifiers = append(identifiers, &ResourceIdentifier{
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		Name:             "test-cr2",
	})

	for _, object := range objects {
		err := k8sClient.Create(context.Background(), object)
		require.NoError(t, err)
	}

	return k8sClient, identifiers
}
