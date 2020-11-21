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
	"hash/fnv"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxTotalClashes = 10

type scenario struct {
	timeout     time.Duration
	identifiers Identifiers
}

type scenarioProvider struct {
	exclusions          []Identifiers
	predefinedScenarios []*scenario
	randomTimeout       time.Duration
	random              int64
	minResources        int64
	maxResources        int64
	computed            sync.Once
	computedScenarios   []*scenario
}

func (sp *scenarioProvider) scenario(system System) (*scenario, bool) {
	sp.computed.Do(func() {
		sp.computedScenarios = append(sp.computedScenarios, sp.predefinedScenarios...)

		if sp.random <= 0 {
			return
		}

		totalClashes := 0
		exclusionHashes := generateExclusionHashes(sp.exclusions)

		for _, predefinedScenario := range sp.predefinedScenarios {
			exclusionHashes = append(exclusionHashes, generateIdentifiersHash(predefinedScenario.identifiers))
		}

		allIdentifiers := system.Identifiers()

		for i := int64(0); i < sp.random; i++ {
			if totalClashes == maxTotalClashes {
				panic("too many clashes with exclusions")
			}
			noOfIdentifiers := rand.Int63n(sp.maxResources-sp.minResources) + sp.minResources

			var identifiers Identifiers

			for j := int64(0); j < noOfIdentifiers; j++ {
				identIdx := rand.Intn(len(allIdentifiers))
				identifiers = append(identifiers, allIdentifiers[identIdx])
			}

			identifierHash := generateIdentifiersHash(identifiers)
			if exists(exclusionHashes, identifierHash) {
				i--
				totalClashes++
				continue
			}

			exclusionHashes = append(exclusionHashes, generateIdentifiersHash(identifiers))
			sp.computedScenarios = append(sp.computedScenarios, &scenario{
				timeout:     sp.randomTimeout,
				identifiers: identifiers,
			})
		}
	})

	if len(sp.computedScenarios) > 0 {
		scenario := sp.computedScenarios[0]
		sp.computedScenarios = sp.computedScenarios[1:len(sp.computedScenarios)]

		return scenario, true
	}

	return nil, false
}

func generateExclusionHashes(exclusions []Identifiers) []string {
	var exclusionHashes []string

	for _, identifiers := range exclusions {
		hash := generateIdentifiersHash(identifiers)
		exclusionHashes = append(exclusionHashes, hash)
	}

	return exclusionHashes
}

func generateIdentifiersHash(identifiers Identifiers) string {
	var ids []string

	for _, identifier := range identifiers {
		ids = append(ids, string(identifier.ID()))
	}

	sort.Strings(ids)

	idsLiteral := strings.Join(ids, "")

	hasher := fnv.New64a()
	// fnv write operation always succeeds
	_, _ = hasher.Write([]byte(idsLiteral))

	return strconv.FormatUint(hasher.Sum64(), 10)
}

func exists(excludedHashes []string, hash string) bool {
	for _, excludedHash := range excludedHashes {
		if hash == excludedHash {
			return true
		}
	}

	return false
}
