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

import (
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const (
	systemsKey          = "systems"
	systemKey           = "system"
	nameKey             = "name"
	typeKey             = "type"
	readyKey            = "ready"
	destroyKey          = "destroy"
	scenariosKey        = "scenarios"
	exclusionsKey       = "exclusions"
	randomKey           = "random"
	minResourcesKey     = "minResources"
	maxResourcesKey     = "maxResources"
	timeoutKey          = "timeout"
	sectionUndefinedErr = "'%s' section not defined"
	strTypeErrMsg       = "'%s' field should be of type string"
	defaultTimeout      = 10 * time.Minute
)

// Config represents the input configuration provided to execute chaos scenarios.
type Config struct {
	ready             ReadyCond
	readyTimeout      time.Duration
	systemNames       map[string]string
	systems           map[string]System
	scenarioProviders map[string]*scenarioProvider
}

// NewConfig instantiates default Config struct.
func NewConfig() *Config {
	return &Config{
		systemNames:       make(map[string]string),
		systems:           make(map[string]System),
		scenarioProviders: make(map[string]*scenarioProvider),
	}
}

func (c *Config) System(name string) System {
	return c.systems[name]
}

// Parse parses the input configuration and populates the Config struct.
func (c *Config) Parse(conf []byte) error {
	yamlDefinition := make(map[string]interface{})
	if err := yaml.Unmarshal(conf, &yamlDefinition); err != nil {
		return errors.Wrap(err, "failed to unmarshall configuration")
	}

	systems, ok := yamlDefinition[systemsKey]
	if !ok {
		return errors.Errorf(sectionUndefinedErr, systemsKey)
	}

	if err := c.parseSystems(systems); err != nil {
		return errors.Wrapf(err, "failed to parse section '%s'", systemsKey)
	}

	ready, ok := yamlDefinition[readyKey]
	if !ok {
		return errors.Errorf(sectionUndefinedErr, readyKey)
	}

	if err := c.parseReady(ready); err != nil {
		return errors.Wrapf(err, "failed to parse section '%s'", readyKey)
	}

	destroy, ok := yamlDefinition[destroyKey]
	if !ok {
		return errors.Errorf(sectionUndefinedErr, destroyKey)
	}

	if err := c.parseDestroy(destroy); err != nil {
		return errors.Wrapf(err, "failed to parse section '%s'", destroyKey)
	}

	return nil
}

func (c *Config) parseSystems(systems interface{}) error {
	systemConfigs, ok := systems.([]interface{})
	if !ok {
		return errors.Errorf("'%s' section should be of type array", systemsKey)
	}

	for _, systemConfig := range systemConfigs {
		system, ok := systemConfig.(map[string]interface{})
		if !ok {
			return errors.Errorf("malformed system configuration %v", systemConfig)
		}

		name, ok := system[nameKey]
		if !ok {
			return errors.Errorf("'%s' field is mandatory in system", nameKey)
		}

		systemName, ok := name.(string)
		if !ok {
			return errors.Errorf(strTypeErrMsg, nameKey)
		}

		sysType, ok := system[typeKey]
		if !ok {
			return errors.Errorf("'%s' field is mandatory in system", typeKey)
		}

		systemType, ok := sysType.(string)
		if !ok {
			return errors.Errorf(strTypeErrMsg, typeKey)
		}

		systemCreator, ok := availableSystems[systemType]
		if !ok {
			return errors.Errorf("unidentified system type '%s'", systemType)
		}

		typedSystem := systemCreator()

		err := typedSystem.Parse(system)
		if err != nil {
			return errors.Wrapf(err, "failed to parse system '%s' of type '%s'", systemName, systemType)
		}

		c.systems[systemName] = typedSystem
		c.systemNames[systemName] = systemType
	}

	return nil
}

func (c *Config) parseReady(ready interface{}) error {
	readyConf, ok := ready.(map[string]interface{})
	if !ok {
		return errors.Errorf("'%s' section should be of type map", readyKey)
	}

	var err error

	timeout := defaultTimeout
	if timeoutValue, ok := readyConf[timeoutKey]; ok {
		timeout, err = parseDuration(timeoutKey, timeoutValue)
		if err != nil {
			return errors.Wrap(err, "failed to parse ready timeout")
		}
	}

	c.readyTimeout = timeout

	for key, parser := range readyParsers {
		if _, ok := readyConf[key]; !ok {
			continue
		}

		readyParser := parser(c)

		readyCond, err := readyParser.Parse(readyConf)
		if err != nil {
			return errors.Wrapf(err, "failed to parse ready section using '%s' ready parser", key)
		}

		c.ready = readyCond

		return nil
	}

	return errors.New("unidentified ready section")
}

func (c *Config) parseDestroy(destroy interface{}) error {
	destroyConf, ok := destroy.(map[string]interface{})
	if !ok {
		return errors.Errorf("'%s' section should be of type map", destroyKey)
	}

	if exclusions, ok := destroyConf[exclusionsKey]; ok {
		if err := c.parseExclusions(exclusions); err != nil {
			return errors.Wrap(err, "failed to wrap exclusions")
		}
	}

	scenarios, ok := destroyConf[scenariosKey]
	if !ok {
		return errors.Errorf("'%s' section is mandatory in '%s", scenariosKey, destroyKey)
	}

	if err := c.parseScenarios(scenarios); err != nil {
		return errors.Wrap(err, "failed to parse scenarios")
	}

	return nil
}

func (c *Config) parseExclusions(exclusions interface{}) error {
	exclusionConfigs, ok := exclusions.([]interface{})
	if !ok {
		return errors.Errorf("'%s' section should be of type array", scenariosKey)
	}

	for _, exclusionConf := range exclusionConfigs {
		exclusion, ok := exclusionConf.(map[string]interface{})
		if !ok {
			return errors.Errorf("malformed exclusion configuration %v", exclusionConf)
		}

		systemNameValue, ok := exclusion[systemKey]
		if !ok {
			return errors.Errorf("'%s' field is mandatory in exclusion", systemKey)
		}

		systemName, ok := systemNameValue.(string)
		if !ok {
			return errors.Errorf(strTypeErrMsg, typeKey)
		}

		systemType, ok := c.systemNames[systemName]
		if !ok {
			return errors.Errorf("system '%s' referenced in a exclusion is not defined", systemName)
		}

		destroyer, ok := availableDestroyers[systemType]
		if !ok {
			return errors.Errorf("destroyer not available for system '%s' of type '%s'", systemName, systemType)
		}

		identifiers, err := destroyer.ParseDestroySection(exclusion)
		if err != nil {
			return errors.Wrapf(err, "failed to parse exclusion '%v' for system type '%s'", exclusion, systemType)
		}

		scenarioProvider := c.scenarioProvider(systemName)
		scenarioProvider.exclusions = append(scenarioProvider.exclusions, identifiers)
	}

	return nil
}

func (c *Config) parseScenarios(scenarios interface{}) error {
	scenariosConfigs, ok := scenarios.([]interface{})
	if !ok {
		return errors.Errorf("'%s' section should be of type array", scenariosKey)
	}

	for _, scenarioConf := range scenariosConfigs {
		scenarioSection, ok := scenarioConf.(map[string]interface{})
		if !ok {
			return errors.Errorf("malformed scenario configuration %v", scenarioConf)
		}

		systemNameValue, ok := scenarioSection[systemKey]
		if !ok {
			return errors.Errorf("'%s' field is mandatory in scenario", systemKey)
		}

		systemName, ok := systemNameValue.(string)
		if !ok {
			return errors.Errorf(strTypeErrMsg, typeKey)
		}

		systemType, ok := c.systemNames[systemName]
		if !ok {
			return errors.Errorf("system '%s' referenced in a scenario is not defined", systemName)
		}

		if _, ok := scenarioSection[randomKey]; ok {
			if err := c.parseRandomScenario(systemName, scenarioSection); err != nil {
				return errors.Wrapf(err, "failed to parse 'random' scenario for system '%s'", systemName)
			}

			return nil
		}

		var err error

		timeout := defaultTimeout
		if timeoutValue, ok := scenarioSection[timeoutKey]; ok {
			timeout, err = parseDuration(timeoutKey, timeoutValue)
			if err != nil {
				return err
			}
		}

		destroyer, ok := availableDestroyers[systemType]
		if !ok {
			return errors.Errorf("destroyer not available for system '%s' of type '%s'", systemName, systemType)
		}

		identifiers, err := destroyer.ParseDestroySection(scenarioSection)
		if err != nil {
			return errors.Wrapf(err, "failed to parse scenario '%v' for system type '%s'", scenarioSection, systemType)
		}

		chaosScenario := &scenario{
			timeout:     timeout,
			identifiers: identifiers,
		}

		scenarioProvider := c.scenarioProvider(systemName)
		scenarioProvider.predefinedScenarios = append(scenarioProvider.predefinedScenarios, chaosScenario)
	}

	return nil
}

func (c *Config) parseRandomScenario(systemName string, scenario map[string]interface{}) error {
	randomValue, ok := (scenario[randomKey]).(float64)
	if !ok {
		return errors.Errorf("'%s' field should be of type float", randomKey)
	}

	random := int64(randomValue)

	var err error

	timeout := defaultTimeout
	if timeoutValue, ok := scenario[timeoutKey]; ok {
		timeout, err = parseDuration(timeoutKey, timeoutValue)
		if err != nil {
			return err
		}
	}

	minimum := int64(1)
	if minValue, ok := scenario[minResourcesKey]; ok {
		minResources, ok := minValue.(float64)
		if !ok {
			return errors.Errorf("'%s' field should be of type int", minResourcesKey)
		}

		minimum = int64(minResources)
	}

	maximum := int64(5)
	if maxValue, ok := scenario[maxResourcesKey]; ok {
		maxResources, ok := maxValue.(float64)
		if !ok {
			return errors.Errorf("'%s' field should be of type int", maxResourcesKey)
		}

		maximum = int64(maxResources)
	}

	scenarioProvider := c.scenarioProvider(systemName)
	scenarioProvider.random = random
	scenarioProvider.minResources = minimum
	scenarioProvider.maxResources = maximum
	scenarioProvider.randomTimeout = timeout

	return nil
}

func (c *Config) scenarioProvider(systemName string) *scenarioProvider {
	if scenarioProvider, ok := c.scenarioProviders[systemName]; ok {
		return scenarioProvider
	}

	scenarioProvider := &scenarioProvider{}
	c.scenarioProviders[systemName] = scenarioProvider

	return scenarioProvider
}

func parseDuration(fieldName string, value interface{}) (time.Duration, error) {
	durationValue, ok := value.(string)
	if !ok {
		return 0, errors.Errorf(strTypeErrMsg, fieldName)
	}

	duration, err := time.ParseDuration(durationValue)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse duration for field '%s'", fieldName)
	}

	return duration, nil
}
