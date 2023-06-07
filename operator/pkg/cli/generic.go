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

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"

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

func getNextK8sResource(dec *yaml.YAMLOrJSONDecoder) (*k8sResource, bool, error) {
	unstructuredObject := &unstructured.Unstructured{}
	if err := dec.Decode(unstructuredObject); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, false, nil
		}
		return nil, false, err
	}
	// Get the resource kind and original bytes
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

func (sc *SchedulerClient) Load(data []byte) ([]proto.Message, error) {
	var protos []proto.Message
	dec := createDecoder(data)
	for {
		resource, keepGoing, err := getNextK8sResource(dec)
		if err != nil {
			return nil, err
		}
		if !keepGoing {
			return protos, nil
		}
		var res proto.Message
		switch resource.kind {
		case model:
			res, err = sc.LoadModel(resource.data)
		case pipeline:
			res, err = sc.LoadPipeline(resource.data)
		case experiment:
			res, err = sc.StartExperiment(resource.data)
		}
		if err != nil {
			return nil, err
		}
		protos = append(protos, res)
	}
}

func (sc *SchedulerClient) Unload(data []byte) ([]proto.Message, error) {
	var protos []proto.Message
	dec := createDecoder(data)
	for {
		resource, keepGoing, err := getNextK8sResource(dec)
		if err != nil {
			return nil, err
		}
		if !keepGoing {
			return protos, err
		}
		var res proto.Message
		switch resource.kind {
		case model:
			res, err = sc.UnloadModel("", resource.data)
		case pipeline:
			res, err = sc.UnloadPipeline("", resource.data)
		case experiment:
			res, err = sc.StopExperiment("", resource.data)
		}
		if err != nil {
			return nil, err
		}
		protos = append(protos, res)
	}
}

func (sc *SchedulerClient) Status(data []byte, wait bool, timeout int64) ([]proto.Message, error) {
	var protos []proto.Message
	dec := createDecoder(data)
	for {
		resource, keepGoing, err := getNextK8sResource(dec)
		if err != nil {
			return nil, err
		}
		if !keepGoing {
			return protos, err
		}
		waitCondition := ""
		var res proto.Message
		switch resource.kind {
		case model:
			if wait {
				waitCondition = "ModelAvailable"
			}
			res, err = sc.ModelStatus(resource.name, waitCondition, timeout)
		case pipeline:
			if wait {
				waitCondition = "PipelineReady"
			}
			res, err = sc.PipelineStatus(resource.name, waitCondition, timeout)
		case experiment:
			res, err = sc.ExperimentStatus(resource.name, wait, timeout)
		}
		if err != nil {
			return nil, err
		}
		protos = append(protos, res)
	}
}
