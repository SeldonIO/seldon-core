/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package common

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/imdario/mergo"
	json "github.com/json-iterator/go"
	v1 "k8s.io/api/core/v1"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
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

// TODO only containers are handled correctly for merging via the name of the container. Need to hande other slices
func MergePodSpecs(serverConfigPodSpec *v1.PodSpec, override *mlopsv1alpha1.PodSpec) (*v1.PodSpec, error) {
	dst := serverConfigPodSpec.DeepCopy()
	if override != nil {
		v1PodSpecOverride, err := override.ToV1PodSpec()
		if err != nil {
			return nil, err
		}

		// remove and copy existing containers
		existingContainers := serverConfigPodSpec.Containers
		err = mergo.Merge(dst, v1PodSpecOverride, mergo.WithOverride, mergo.WithAppendSlice)
		if err != nil {
			return nil, err
		}

		// merge containers
		updatedConatiners, err := MergeContainers(existingContainers, override.Containers)
		if err != nil {
			return nil, err
		}
		dst.Containers = updatedConatiners

		return dst, nil
	} else {
		return dst, nil
	}
}

// Allow containers to be overridden. As containers are keys by name we need to merge by this key.
func MergeContainers(existing []v1.Container, overrides []v1.Container) ([]v1.Container, error) {
	var containersNew []v1.Container
	for _, containerOverride := range overrides {
		found := false
		for _, containerExisting := range existing {
			if containerOverride.Name == containerExisting.Name {
				found = true
				err := mergo.Merge(&containerExisting, containerOverride, mergo.WithOverride, mergo.WithAppendSlice)
				if err != nil {
					return nil, err
				}
				containersNew = append(containersNew, containerExisting)
			}
		}
		if !found {
			containersNew = append(containersNew, containerOverride)
		}
	}
	for _, containerExisting := range existing {
		found := false
		for _, containerOverride := range overrides {
			if containerExisting.Name == containerOverride.Name {
				found = true
			}
		}
		if !found {
			containersNew = append(containersNew, containerExisting)
		}
	}
	return containersNew, nil
}
