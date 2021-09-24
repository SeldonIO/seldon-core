//   Copyright Steve Sloka 2021
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package processor

import (
	"github.com/seldonio/seldon-core/scheduler/apis/v1alpha1"
	"gopkg.in/yaml.v2"
)

// parseYaml takes in a yaml envoy config and returns a typed version
func parseSeldonYaml(data []byte) (*v1alpha1.SeldonConfig, error) {
	var config v1alpha1.SeldonConfig

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
