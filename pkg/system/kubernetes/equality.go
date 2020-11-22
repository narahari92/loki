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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	status          = "status"
	conditions      = "conditions"
	conditionType   = "type"
	conditionStatus = "status"
)

var semanticEquality = conversion.EqualitiesOrDie(
	func(a, b resource.Quantity) bool {
		// Ignore formatting, only care that numeric value stayed the same.
		// Uninitialized quantities are equivalent to 0 quantities.
		return a.Cmp(b) == 0
	},
	func(a, b metav1.MicroTime) bool {
		// Don't compare time
		return true
	},
	func(a, b metav1.Time) bool {
		// Dont compare time
		return true
	},
	func(a, b labels.Selector) bool {
		return a.String() == b.String()
	},
	func(a, b fields.Selector) bool {
		return a.String() == b.String()
	},
)

func isEqual(desired, actual *unstructured.Unstructured) bool {
	if ok := semanticEquality.DeepEqual(desired.GetLabels(), actual.GetLabels()); !ok {
		return false
	}

	desiredStatusValue, ok, err := unstructured.NestedFieldNoCopy(desired.Object, status)
	if err != nil || !ok {
		return true
	}

	actualStatusValue, ok, err := unstructured.NestedFieldNoCopy(actual.Object, status)
	if err != nil || !ok {
		// return false because desired has status but actual don't
		return false
	}

	desiredStatus, ok := desiredStatusValue.(map[string]interface{})
	if !ok {
		if _, ok := actualStatusValue.(map[string]interface{}); !ok {
			return semanticEquality.DeepEqual(desiredStatusValue, actualStatusValue)
		}

		return false
	}

	actualStatus := actualStatusValue.(map[string]interface{})

	_, ok, err = unstructured.NestedFieldNoCopy(desiredStatus, conditions)
	if err == nil && ok {
		return compareConditions(desiredStatus, actualStatus)
	}

	if ok := semanticEquality.DeepEqual(desiredStatus, actualStatus); !ok {
		return false
	}

	return compareConditions(desiredStatus, actualStatus)
}

func compareConditions(desiredStatus, actualStatus map[string]interface{}) bool {
	desiredConditionsValue, ok, err := unstructured.NestedFieldNoCopy(desiredStatus, conditions)
	if err != nil || !ok {
		return true
	}

	actualConditionsValue, ok, err := unstructured.NestedFieldNoCopy(actualStatus, conditions)
	if err != nil || !ok {
		// return false because desired has status but actual don't
		return false
	}

	desiredConditions, ok := desiredConditionsValue.([]interface{})
	if !ok {
		return false
	}

	desiredConditonsByType := make(map[interface{}]map[string]interface{})

	for _, desiredConditionValue := range desiredConditions {
		desiredCondition, ok := desiredConditionValue.(map[string]interface{})
		if !ok {
			return false
		}

		desiredConditonsByType[desiredCondition[conditionType]] = desiredCondition
	}

	actualConditions, ok := actualConditionsValue.([]interface{})
	if !ok {
		return false
	}

	actualConditonsByType := make(map[interface{}]map[string]interface{})

	for _, actualConditonValue := range actualConditions {
		actualCondition, ok := actualConditonValue.(map[string]interface{})
		if !ok {
			return false
		}

		actualConditonsByType[actualCondition[conditionType]] = actualCondition
	}

	for key, desiredValue := range desiredConditonsByType {
		actualValue, ok := actualConditonsByType[key]
		if !ok {
			return false
		}

		if ok := semanticEquality.DeepEqual(desiredValue[conditionStatus], actualValue[conditionStatus]); !ok {
			return false
		}

		delete(actualConditonsByType, key)
	}

	if len(actualConditonsByType) != 0 {
		return false
	}

	return true
}
