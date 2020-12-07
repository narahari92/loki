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

package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/narahari92/loki/pkg/audit"
	"github.com/narahari92/loki/pkg/loki"
	"github.com/narahari92/loki/pkg/rego"
	"github.com/narahari92/loki/pkg/system/kubernetes"
)

func main() {
	logger := logrus.New()

	if err := run(context.Background(), logger); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger logrus.FieldLogger) error {
	registerDependencies()

	configFile := flag.String("config", "", "configuration yaml for execution")
	reportLocation := flag.String("report", "", "location where the report file will be created")

	flag.Parse()

	configuration, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read contents of configuration file '%s'", *configFile)
	}

	config := loki.NewConfig()

	if err := config.Parse(configuration); err != nil {
		return errors.Wrap(err, "failed to parse configuration")
	}

	chaosMaker := loki.ChaosMaker{
		Config:      config,
		FieldLogger: logger,
		Reporter:    &audit.Reporter{},
	}

	defer func() {
		if reportLocation == nil || *reportLocation == "" {
			return
		}

		file, err := os.Create(*reportLocation)
		if err != nil {
			logger.WithError(err).Errorf("failed to create report file %s", *reportLocation)
		}

		if err = chaosMaker.Reporter.Report(file); err != nil {
			logger.WithError(err).Errorf("failed to write report into file %s", *reportLocation)
		}

		if err = file.Close(); err != nil {
			logger.WithError(err).Errorf("failed to close report file %s", *reportLocation)
		}
	}()

	if err := chaosMaker.CreateChaos(ctx); err != nil {
		return errors.Wrap(err, "failure in chaos")
	}

	return nil
}

func registerDependencies() {
	const afterKey = "after"

	kubernetes.Register()

	loki.RegisterReadyParser(afterKey, loki.AfterParser)
	rego.RegisterReadyParser()

}
