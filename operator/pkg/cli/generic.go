package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type seldonKind int64

const (
	Undefined seldonKind = iota
	model
	pipeline
	experiment
)

var kindMap = map[string]seldonKind{
	"Model":      model,
	"Pipeline":   pipeline,
	"Experiment": experiment,
}

type k8sResource struct {
	kind seldonKind
	name string
	data []byte
}

func createDecoder(data []byte) *yaml.YAMLOrJSONDecoder {
	var reader io.Reader
	if len(data) > 0 {
		reader = bytes.NewReader(data)
	} else {
		reader = io.Reader(os.Stdin)
	}
	dec := yaml.NewYAMLOrJSONDecoder(reader, 10)
	return dec
}

func getNextResource(dec *yaml.YAMLOrJSONDecoder) (*k8sResource, bool, error) {
	unstructuredObject := &unstructured.Unstructured{}
	if err := dec.Decode(unstructuredObject); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, false, nil
		}
		return nil, false, err
	}
	// Get the resource kind nad original bytes
	gvk := unstructuredObject.GroupVersionKind()
	data, err := unstructuredObject.MarshalJSON()
	if err != nil {
		return nil, false, err
	}
	if kind, ok := kindMap[gvk.Kind]; ok {
		return &k8sResource{
			kind: kind,
			name: unstructuredObject.GetName(),
			data: data,
		}, true, nil
	} else {
		return nil, false, fmt.Errorf("Unknown Seldon Kind %s - only Model, Pipeline and Experiment allowed", gvk.Kind)
	}
}

func (sc *SchedulerClient) Load(data []byte, showRequest bool, showResponse bool) error {
	dec := createDecoder(data)
	for {
		resource, keepGoing, err := getNextResource(dec)
		if err != nil {
			return err
		}
		if !keepGoing {
			return err
		}
		switch resource.kind {
		case model:
			err = sc.LoadModel(resource.data, showRequest, showResponse)
		case pipeline:
			err = sc.LoadPipeline(resource.data, showRequest, showResponse)
		case experiment:
			err = sc.StartExperiment(resource.data, showRequest, showResponse)
		}
		if err != nil {
			return err
		}
	}
}

func (sc *SchedulerClient) Unload(data []byte, showRequest bool, showResponse bool) error {
	dec := createDecoder(data)
	for {
		resource, keepGoing, err := getNextResource(dec)
		if !keepGoing {
			return err
		}
		switch resource.kind {
		case model:
			err = sc.UnloadModel("", resource.data, showRequest, showResponse)
		case pipeline:
			err = sc.UnloadPipeline("", resource.data, showRequest, showResponse)
		case experiment:
			err = sc.StopExperiment("", resource.data, showRequest, showResponse)
		}
		if err != nil {
			return err
		}
	}
}

func (sc *SchedulerClient) Status(data []byte, showRequest bool, showResponse bool, wait bool) error {
	dec := createDecoder(data)
	for {
		resource, keepGoing, err := getNextResource(dec)
		if !keepGoing {
			return err
		}
		waitCondition := ""
		switch resource.kind {
		case model:
			if wait {
				waitCondition = "ModelAvailable"
			}
			err = sc.ModelStatus(resource.name, showRequest, showResponse, waitCondition)
		case pipeline:
			if wait {
				waitCondition = "PipelineReady"
			}
			err = sc.PipelineStatus(resource.name, showRequest, showResponse, waitCondition)
		case experiment:
			err = sc.ExperimentStatus(resource.name, showRequest, showResponse, wait)
		}
		if err != nil {
			return err
		}
	}
}
