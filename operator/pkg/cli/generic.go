package cli

import (
	"bytes"
	"errors"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
)

func createDecoder(data []byte) *yaml2.YAMLOrJSONDecoder {
	var reader io.Reader
	if len(data) > 0 {
		reader = bytes.NewReader(data)
	} else {
		reader = io.Reader(os.Stdin)
	}
	dec := yaml2.NewYAMLOrJSONDecoder(reader, 10)
	return dec
}

func getNextResource(dec *yaml2.YAMLOrJSONDecoder) (string, string, []byte, bool, error) {
	unstructuredObject := &unstructured.Unstructured{}
	if err := dec.Decode(unstructuredObject); err != nil {
		if errors.Is(err, io.EOF) {
			return "", "", nil, false, nil
		}
		return "", "", nil, false, err
	}
	// Get the resource kind nad original bytes
	gvk := unstructuredObject.GroupVersionKind()
	data, err := unstructuredObject.MarshalJSON()
	if err != nil {
		return "", "", nil, false, err
	}
	return gvk.Kind, unstructuredObject.GetName(), data, true, nil
}

func (sc *SchedulerClient) Load(data []byte, showRequest bool, showResponse bool) error {
	dec := createDecoder(data)
	for {
		kind, _, data, keepGoing, err := getNextResource(dec)
		if !keepGoing {
			return err
		}
		switch kind {
		case "Model":
			err = sc.LoadModel(data, showRequest, showResponse)
		case "Pipeline":
			err = sc.LoadPipeline(data, showRequest, showResponse)
		case "Experiment":
			err = sc.StartExperiment(data, showRequest, showResponse)
		}
		if err != nil {
			return err
		}
	}
}

func (sc *SchedulerClient) Unload(data []byte, showRequest bool, showResponse bool) error {
	dec := createDecoder(data)
	for {
		kind, _, data, keepGoing, err := getNextResource(dec)
		if !keepGoing {
			return err
		}
		switch kind {
		case "Model":
			err = sc.UnloadModel("", data, showRequest, showResponse)
		case "Pipeline":
			err = sc.UnloadPipeline("", data, showRequest, showResponse)
		case "Experiment":
			err = sc.StopExperiment("", data, showRequest, showResponse)
		}
		if err != nil {
			return err
		}
	}
}

func (sc *SchedulerClient) Status(data []byte, showRequest bool, showResponse bool, wait bool) error {
	dec := createDecoder(data)
	for {
		kind, name, _, keepGoing, err := getNextResource(dec)
		if !keepGoing {
			return err
		}
		waitCondition := ""
		switch kind {
		case "Model":
			if wait {
				waitCondition = "ModelAvailable"
			}
			err = sc.ModelStatus(name, showRequest, showResponse, waitCondition)
		case "Pipeline":
			if wait {
				waitCondition = "PipelineReady"
			}
			err = sc.PipelineStatus(name, showRequest, showResponse, waitCondition)
		case "Experiment":
			err = sc.ExperimentStatus(name, showRequest, showResponse, wait)
		}
		if err != nil {
			return err
		}
	}
}
