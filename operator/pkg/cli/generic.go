/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

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
			res, err = sc.UnloadModel(resource.name, resource.data)
		case pipeline:
			res, err = sc.UnloadPipeline(resource.name, resource.data)
		case experiment:
			res, err = sc.StopExperiment(resource.name, resource.data)
		}
		if err != nil {
			return nil, err
		}
		protos = append(protos, res)
	}
}

func (sc *SchedulerClient) Status(data []byte, wait bool, timeout time.Duration) ([]proto.Message, error) {
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
