package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	resourceModelKind      = "Model"
	resourcePipelineKind   = "Pipeline"
	resourceExperimentKind = "Experiment"
	resourceServerKind     = "Server"
	InferResponseIDField   = "id"
)

type SeldonResourceMeta struct {
	name         string
	gvk          schema.GroupVersionKind
	unstructured *unstructured.Unstructured
	obj          client.Object
}

func getResourceFromKind(kind string) (client.Object, error) {
	switch kind {
	case resourceModelKind:
		return &v1alpha1.Model{}, nil
	case resourcePipelineKind:
		return &v1alpha1.Pipeline{}, nil
	case resourceExperimentKind:
		return &v1alpha1.Experiment{}, nil
	case resourceServerKind:
		return &v1alpha1.Server{}, nil
	default:
		return nil, fmt.Errorf("Unknown Kind %s", kind)
	}
}

func getResource(filename string) (*SeldonResourceMeta, error) {
	dat, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(dat)
	dec := yaml.NewYAMLOrJSONDecoder(reader, 10)
	unstructuredObject := &unstructured.Unstructured{}
	if err := dec.Decode(unstructuredObject); err != nil {
		return nil, err
	}
	obj, err := getResourceFromKind(unstructuredObject.GetKind())
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObject.UnstructuredContent(), obj)
	if err != nil {
		return nil, err
	}

	return &SeldonResourceMeta{
		name:         unstructuredObject.GetName(),
		gvk:          unstructuredObject.GroupVersionKind(),
		unstructured: unstructuredObject,
		obj:          obj,
	}, nil
}

func getInferRequestProtocol(request []byte) (cli.InferProtocol, error) {
	var data map[string]interface{}
	err := json.Unmarshal(request, &data)
	if err != nil {
		return cli.InferUnknown, err
	}
	if inputs, ok := data["inputs"]; ok {
		inputList := inputs.([]interface{})
		if len(inputList) >= 1 {
			input := inputList[0].(map[string]interface{})
			if _, ok := input["data"]; ok {
				return cli.InferRest, nil
			} else if _, ok := input["contents"]; ok {
				return cli.InferGrpc, nil
			}
		}
	}
	return cli.InferUnknown, fmt.Errorf("cannot decode infer request payload as rest or grpc json proto")
}

func removeIdFromResponse(response []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(response, &data)
	if err != nil {
		return nil, err
	}
	if _, ok := data[InferResponseIDField]; ok {
		delete(data, InferResponseIDField)
		responseUpdated, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return responseUpdated, nil
	} else {
		return response, nil
	}
}
