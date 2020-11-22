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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestEquality(t *testing.T) {
	pod1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"generation": 1,
				"labels": map[string]string{
					"isPod": "true",
				},
				"annotations": map[string]string{
					"test": "true",
				},
			},
			"spec": map[string]interface{}{
				"image": "someimage",
			},
			"status": map[string]interface{}{
				"timestamp": metav1.Now(),
			},
		},
	}

	pod2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"generation": 2,
				"labels": map[string]string{
					"isPod": "true",
				},
				"annotations": map[string]string{
					"test": "true",
				},
			},
			"spec": map[string]interface{}{
				"image": "someimage",
			},
			"status": map[string]interface{}{
				"timestamp": metav1.Time{Time: metav1.Now().Add(1 * time.Hour)},
			},
		},
	}

	equality := isEqual(pod1, pod2)
	require.Equal(t, true, equality)
}
