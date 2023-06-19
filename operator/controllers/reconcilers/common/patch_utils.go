/*
Copyright 2023 Seldon Technologies Ltd.

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

package common

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	json "github.com/json-iterator/go"
)

func IgnoreVolumeClaimTemplateVolumeModel() patch.CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteVolumeClaimTemplateFields(current)
		if err != nil {
			return []byte{}, []byte{}, errors.Wrap(err, "could not delete status field from current byte sequence")
		}

		modified, err = deleteVolumeClaimTemplateFields(modified)
		if err != nil {
			return []byte{}, []byte{}, errors.Wrap(err, "could not delete status field from modified byte sequence")
		}

		return current, modified, nil
	}
}

func deleteVolumeClaimTemplateFields(obj []byte) ([]byte, error) {
	resource := map[string]interface{}{}
	err := json.Unmarshal(obj, &resource)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not unmarshal byte sequence")
	}

	if spec, ok := resource["spec"]; ok {
		if spec, ok := spec.(map[string]interface{}); ok {
			if vcts, ok := spec["volumeClaimTemplates"]; ok {
				if vcts, ok := vcts.([]interface{}); ok {
					for _, vct := range vcts {
						if vct, ok := vct.(map[string]interface{}); ok {
							if spec, ok := vct["spec"]; ok {
								if spec, ok := spec.(map[string]interface{}); ok {
									spec["volumeMode"] = ""
								}
							}
						}
					}
				}
			}
		}
	}

	obj, err = json.ConfigCompatibleWithStandardLibrary.Marshal(resource)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not marshal byte sequence")
	}

	return obj, nil
}
