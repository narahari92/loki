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
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/narahari92/loki/pkg/loki"
	"github.com/narahari92/loki/pkg/wait"
)

const (
	kubeconfigKey  = "kubeconfig"
	inclusterKey   = "incluster"
	resourcesKey   = "resources"
	apiVersionKey  = "apiVersion"
	kindKey        = "kind"
	nameKey        = "name"
	namespaceKey   = "namespace"
	strTypeErrMsg  = "'%s' field should be of type string"
	reqFieldErrMsg = "'%s' field is required for kubernetes resource"
)

// System represents a kubernetes system comprising of resources as defined in input configuration.
type System struct {
	kubeconfig          string
	inCluster           bool
	k8sClient           client.Client
	resourceIdentifiers []*ResourceIdentifier
	state               map[ResourceIdentifier]*unstructured.Unstructured
	logger              logrus.FieldLogger
}

// NewSystem instantiates kubernetes System.
func NewSystem() *System {
	return &System{
		state:  make(map[ResourceIdentifier]*unstructured.Unstructured),
		logger: logrus.New().WithField("system", system),
	}
}

// Parse parses the configuration of system given in input configuration.
func (s *System) Parse(systemConfig map[string]interface{}) error {
	if kubecfg, ok := systemConfig[kubeconfigKey]; ok {
		kubeconfig, ok := kubecfg.(string)
		if !ok {
			return errors.Errorf(strTypeErrMsg, kubeconfigKey)
		}

		s.kubeconfig = kubeconfig
	}

	if incluster, ok := systemConfig[inclusterKey]; ok {
		inClusterSystem, ok := incluster.(bool)
		if !ok {
			return errors.Errorf("'%s' field should be of type bool", inclusterKey)
		}

		s.inCluster = inClusterSystem
	}

	if s.kubeconfig == "" && !s.inCluster {
		return errors.Errorf("either '%s' or '%s' as true must be specified", kubeconfigKey, inclusterKey)
	}

	resources, ok := systemConfig[resourcesKey]
	if !ok {
		return errors.Errorf("'%s' field must be defined for kubernetes system", resourcesKey)
	}

	k8sResources, ok := resources.([]interface{})
	if !ok {
		return errors.Errorf("'%s' field should be of type array", resourcesKey)
	}

	identifiers, err := parseResources(k8sResources)
	if err != nil {
		return err
	}

	s.resourceIdentifiers = identifiers

	if err := s.createClient(); err != nil {
		return err
	}

	return nil
}

// Load loads all the kubernetes resources defined in system of input configuration and stores it in memory. This will be
// used in validation during chaos testing.
func (s *System) Load(ctx context.Context) error {
	for _, resourceIdentifier := range s.resourceIdentifiers {
		if resourceIdentifier.Name != "" {
			object := &unstructured.Unstructured{}
			object.SetGroupVersionKind(resourceIdentifier.GroupVersionKind)

			if err := s.k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: resourceIdentifier.Namespace, Name: resourceIdentifier.Name},
				object,
			); err != nil {
				return errors.Wrap(err, "failed to get kubernetes resource")
			}

			s.state[*resourceIdentifier] = object
			continue
		}

		objects := &unstructured.UnstructuredList{}
		objects.SetGroupVersionKind(resourceIdentifier.GroupVersionKind)

		if err := s.k8sClient.List(
			ctx,
			objects,
			&client.ListOptions{Namespace: resourceIdentifier.Namespace},
		); err != nil {
			return errors.Wrap(err, "failed to list kubernetes resource")
		}

		for _, object := range objects.Items {
			object := object
			gv, err := schema.ParseGroupVersion(object.GetAPIVersion())
			if err != nil {
				return errors.Wrap(err, "failed to parse group version")
			}

			resIdent := ResourceIdentifier{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   gv.Group,
					Version: gv.Version,
					Kind:    object.GetKind(),
				},
				Namespace: object.GetNamespace(),
				Name:      object.GetName(),
			}

			s.state[resIdent] = &object
		}
	}

	return nil
}

// Validate validates whether the system is in desired state or not by comparing kubernetes resources at current time with
// that loaded by Load function.
func (s *System) Validate(ctx context.Context) (bool, error) {
	backoff := wait.MinMaxBackoff{
		Min: 250 * time.Millisecond,
		Max: 500 * time.Millisecond,
	}

	for identifier, resource := range s.state {
		object := &unstructured.Unstructured{}
		object.SetGroupVersionKind(identifier.GroupVersionKind)

		if err := s.k8sClient.Get(
			ctx,
			types.NamespacedName{Namespace: identifier.Namespace, Name: identifier.Name},
			object); err != nil {
			return false, errors.Wrap(err, "failed to get kubernetes resource")
		}

		ok := isEqual(resource, object)
		if !ok {
			s.logger.Warnf("resource '%s' of kind '%s' in '%s' namespace didn't reach desired state",
				identifier.Name, identifier.Kind, identifier.Namespace)
			s.logger.Debugf("resource %#v\ndidn't match with\n%#v", resource, object)
			return ok, nil
		}

		time.Sleep(backoff.Step())
	}

	return true, nil
}

// Identifiers return Identifier values of all resources in the kubernetes system.
func (s *System) Identifiers() loki.Identifiers {
	var identifiers loki.Identifiers

	for identifier := range s.state {
		identifier := identifier
		identifiers = append(identifiers, &identifier)
	}

	return identifiers
}

// AsJSON returns the json representation of the state of the kubernetes system. If `reload` is set to `true`, state of the system
// will be reloaded before preparing json representation of system.
func (s *System) AsJSON(ctx context.Context, reload bool) ([]byte, error) {
	if reload {
		if err := s.Load(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to get json representation of system")
		}
	}

	objects := make([]unstructured.Unstructured, 0)

	for _, obj := range s.state {
		objects = append(objects, *obj)
	}

	return json.Marshal(objects)
}

func (s *System) createClient() error {
	var restCfg *rest.Config

	var err error

	if s.kubeconfig != "" {
		kubecfgData, err := ioutil.ReadFile(s.kubeconfig)
		if err != nil {
			return errors.Wrapf(err, "failed to read kubeconfig file '%s'", s.kubeconfig)
		}

		restCfg, err = clientcmd.RESTConfigFromKubeConfig(kubecfgData)
		if err != nil {
			return errors.Wrap(err, "failed to create rest config from kubeconfig")
		}
	}

	if restCfg == nil {
		restCfg, err = rest.InClusterConfig()
		if err != nil {
			return errors.Wrap(err, "failed to get in-cluster rest config")
		}
	}

	k8sClient, err := client.New(restCfg, client.Options{})
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes client")
	}

	s.k8sClient = k8sClient

	return nil
}
