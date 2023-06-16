package utils

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	json "github.com/json-iterator/go"
)

func IgnoreReplicas() patch.CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteReplicas(current)
		if err != nil {
			return []byte{}, []byte{}, errors.Wrap(err, "could not delete replicas field from current byte sequence")
		}

		modified, err = deleteReplicas(modified)
		if err != nil {
			return []byte{}, []byte{}, errors.Wrap(err, "could not delete replicas field from modified byte sequence")
		}

		return current, modified, nil
	}
}

func deleteReplicas(obj []byte) ([]byte, error) {
	resource := map[string]interface{}{}
	err := json.Unmarshal(obj, &resource)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not unmarshal byte sequence")
	}

	if spec, ok := resource["spec"]; ok {
		if spec, ok := spec.(map[string]interface{}); ok {
			if _, ok := spec["replicas"]; ok {
				spec["replicas"] = nil
			}
		}
	}

	obj, err = json.ConfigCompatibleWithStandardLibrary.Marshal(resource)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not marshal byte sequence")
	}

	return obj, nil
}
