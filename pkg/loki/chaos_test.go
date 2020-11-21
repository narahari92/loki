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
	"io/ioutil"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestChaos(t *testing.T) {
	RegisterTestSystem(t)

	conf, err := ioutil.ReadFile("../../testdata/sample-config.yaml")
	require.NoError(t, err)

	configuration := NewConfig()

	err = configuration.Parse([]byte(conf))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	go func(ctx context.Context, system *TestSystem) {
		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(1 * time.Second):
				if _, ok := system.state[TestIdentifier("resource1")]; !ok {
					break
				}

				_, res2Ok := system.state[TestIdentifier("resource2")]
				_, res3Ok := system.state[TestIdentifier("resource3")]
				if !res2Ok && !res3Ok {
					break
				}

				system.state[TestIdentifier("resource2")] = true
				system.state[TestIdentifier("resource3")] = true
				system.state[TestIdentifier("resource4")] = true
			}
		}
	}(ctx, configuration.systems["testing"].(*TestSystem))

	errChan := make(chan error)

	go func() {
		chaosMaker := &ChaosMaker{
			Config:      configuration,
			FieldLogger: logrus.New(),
		}

		err = chaosMaker.CreateChaos(ctx)
		errChan <- err
	}()

	select {
	case <-ctx.Done():
		err = errors.New("timed out waiting for recovery from chaos")

	case err = <-errChan:
	}

	require.NoError(t, err)
}
