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

package rego

import (
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/narahari92/loki/pkg/loki"
)

const (
	regoKey       = "rego"
	systemKey     = "system"
	policyFileKey = "policyFile"
	queryKey      = "query"
)

// ReadyParser parses the ready section to create ReadyCond which identifies readiness based on rego policy.
type ReadyParser struct {
	config *loki.Config
}

// NewReadyParser instantiates ReadyParser.
func NewReadyParser(config *loki.Config) loki.ReadyParser {
	return &ReadyParser{
		config: config,
	}
}

// Parse parses the ready section which uses rego policy and creates ReadyCond.
func (parser *ReadyParser) Parse(readyConf map[string]interface{}) (loki.ReadyCond, error) {
	regoConf, ok := readyConf[regoKey].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("field '%s' should be of type map", regoKey)
	}

	systemNameValue, ok := regoConf[systemKey]
	if !ok {
		return nil, errors.Errorf("field '%s' is mandatory under '%s'", systemKey, regoKey)
	}

	systemName, ok := systemNameValue.(string)
	if !ok {
		return nil, errors.Errorf("field '%s' should be of type string", systemKey)
	}

	system := parser.config.System(systemName)
	if system == nil {
		return nil, errors.Errorf("unidentified system '%s' referenced in readiness validation", systemName)
	}

	policyFileValue, ok := regoConf[policyFileKey]
	if !ok {
		return nil, errors.Errorf("field '%s' is mandatory under '%s'", policyFileKey, regoKey)
	}

	policyFile, ok := policyFileValue.(string)
	if !ok {
		return nil, errors.Errorf("field '%s' should be of type string", policyFileKey)
	}

	policy, err := ioutil.ReadFile(policyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read policy file '%s'", policyFile)
	}

	queryValue, ok := regoConf[queryKey]
	if !ok {
		return nil, errors.Errorf("field '%s' is mandatory under '%s'", queryKey, regoKey)
	}

	query, ok := queryValue.(string)
	if !ok {
		return nil, errors.Errorf("field '%s' should be of type string", queryKey)
	}

	return &ReadyCond{
		query:  query,
		policy: string(policy),
		system: system,
	}, nil
}

// RegisterReadyParser registers rego ReadyParser into loki.
func RegisterReadyParser() {
	loki.RegisterReadyParser(regoKey, NewReadyParser)
}
