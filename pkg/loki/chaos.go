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
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/narahari92/loki/pkg/wait"
)

// ChaosMaker takes Config and executes chaos scenarios both pre-defined and randomly generated ones.
type ChaosMaker struct {
	*Config
	logrus.FieldLogger
}

// CreateChaos executes all the chaos scenarios and exits with error on first scenario which fails to recover and
// get into desired state or returns successfully if all systems get back into desired state from all chaos scenarios.
func (cm *ChaosMaker) CreateChaos(ctx context.Context, opts ...HookOption) error {
	hook := &Hook{}

	for _, opt := range opts {
		opt(hook)
	}

	if err := cm.readyCheck(ctx, hook); err != nil {
		return err
	}

	if err := cm.loadSystems(ctx, hook); err != nil {
		return err
	}

	if hook.preChaos != nil {
		cm.Info("pre chaos hook executing")
		if err := hook.preChaos(ctx); err != nil {
			cm.WithError(err).Warn("pre chaos hook failed")
		}
	}

	if hook.postChaos != nil {
		defer func() {
			cm.Info("post chaos hook executing")
			if err := hook.postChaos(ctx); err != nil {
				cm.WithError(err).Warn("post chaos hook failed")
			}
		}()
	}

	for systemName, provider := range cm.scenarioProviders {
		cm.Infof("creating chaos in '%s' system", systemName)

		systemType := cm.systemNames[systemName]
		system := cm.systems[systemName]

		killerCreator, ok := availableKillers[systemType]
		if !ok {
			errorMsg := "no killer registered for system '%s' of type '%s'"
			cm.Errorf(errorMsg, systemName, systemType)
			return errors.Errorf(errorMsg, systemName, systemType)
		}

		killer, err := killerCreator(system)
		if err != nil {
			errorMsg := "failed to create killer for system '%s' of type '%s'"
			cm.WithError(err).Errorf(errorMsg, systemName, systemType)
			return errors.Wrapf(err, errorMsg, systemName, systemType)
		}

		for {
			scenario, ok := provider.scenario(system)
			if !ok {
				break
			}

			cm.Infof("creating chaos by action:\n%s", scenario.identifiers)
			if err := killer.Kill(ctx, scenario.identifiers...); err != nil {
				errorMsg := "failed to kill identifiers for system %s of type %s"
				cm.WithError(err).Errorf(errorMsg, systemName, systemType)
				return errors.Wrapf(err, errorMsg, systemName, systemType)
			}

			ok, err := wait.ExecuteWithBackoff(
				ctx,
				&wait.ExponentialBackoff{
					Cap:    10 * time.Minute,
					Factor: 2.0,
					Jitter: 0.3,
				},
				system.Validate,
				scenario.timeout,
			)

			if err != nil {
				errorMsg := "failed to validate system '%s'"
				cm.WithError(err).Errorf(errorMsg, systemName)
				return errors.Wrapf(err, errorMsg, systemName)
			}

			if !ok {
				errorMsg := "validation failed. system '%s' didn't reach desired state"
				cm.Errorf(errorMsg, systemName)
				return errors.Errorf(errorMsg, systemName)
			}

			cm.Infof("recovered successfully by chaos by action:\n%s", scenario.identifiers)
		}
	}

	return nil
}

func (cm *ChaosMaker) loadSystems(ctx context.Context, hook *Hook) error {
	if hook.preSystemLoad != nil {
		cm.Info("pre system load hook executing")
		if err := hook.preSystemLoad(ctx); err != nil {
			cm.WithError(err).Warn("pre system load hook failed")
		}
	}

	if hook.postSystemLoad != nil {
		defer func() {
			cm.Info("post system load hook executing")
			if err := hook.postSystemLoad(ctx); err != nil {
				cm.WithError(err).Warn("post system load hook failed")
			}
		}()
	}

	cm.Info("system(s) are being loaded")

	for name, system := range cm.systems {
		if err := system.Load(ctx); err != nil {
			errorMsg := "system '%s' failed to load"
			cm.WithError(err).Errorf(errorMsg, name)
			return errors.Wrapf(err, errorMsg, name)
		}
	}

	cm.Info("system(s) are loaded")

	return nil
}

func (cm *ChaosMaker) readyCheck(ctx context.Context, hook *Hook) error {
	if hook.preReady != nil {
		cm.Info("pre ready hook executing")
		if err := hook.preReady(ctx); err != nil {
			cm.WithError(err).Warn("pre ready hook failed")
		}
	}

	if hook.postReady != nil {
		defer func() {
			cm.Info("post ready hook executing")
			if err := hook.postReady(ctx); err != nil {
				cm.WithError(err).Warn("post ready hook failed")
			}
		}()
	}

	cm.Info("initiating readiness check")

	ok, err := wait.ExecuteWithBackoff(
		ctx,
		&wait.ExponentialBackoff{
			Duration: 1 * time.Second,
			Cap:      10 * time.Minute,
			Factor:   1.5,
			Jitter:   0.7,
		},
		cm.ready.Ready,
		cm.readyTimeout,
	)
	if err != nil {
		errorMsg := "system(s) failed to reach ready state"
		cm.WithError(err).Error(errorMsg)
		return errors.Wrap(err, errorMsg)
	}

	if !ok {
		errorMsg := "system(s) didn't reach ready state"
		cm.Errorf(errorMsg)
		return errors.New(errorMsg)
	}

	cm.Info("system(s) are ready for chaos testing")

	return nil
}
