/*
Copyright 2022 Seldon Technologies Ltd.

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

package kafka

import (
	"fmt"
	"strings"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

const (
	seldonTopicPrefix = "seldon"
	modelTopic        = "model"
	pipelineTopic     = "pipeline"
	errorsTopic       = "errors"
	inputsSuffix      = "inputs"
	outputsSuffix     = "outputs"
	errorsSuffix      = "errors"
	TopicErrorHeader  = "seldon-pipeline-errors"
)

type TopicNamer struct {
	namespace string
}

func NewTopicNamer(namespace string) *TopicNamer {
	if namespace == "" {
		namespace = "default"
	}
	return &TopicNamer{
		namespace: namespace,
	}
}

func (tn *TopicNamer) GetModelErrorTopic() string {
	return fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, errorsTopic, errorsSuffix)
}

func (tn *TopicNamer) GetKafkaModelTopicRegex() string {
	return fmt.Sprintf("^%s.%s.%s.*.%s", seldonTopicPrefix, tn.namespace, modelTopic, inputsSuffix)
}

func (tn *TopicNamer) GetModelNameFromModelInputTopic(topic string) (string, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 5 {
		return "", fmt.Errorf("Wrong number of sections in topic %s. Whas expecting 5 with separator '.'", topic)
	}
	if parts[0] != seldonTopicPrefix || parts[1] != tn.namespace || parts[2] != modelTopic || parts[4] != inputsSuffix {
		return "", fmt.Errorf("Bad topic name %s needs to match %s", topic, tn.GetKafkaModelTopicRegex())
	}
	return parts[3], nil
}

func (tn *TopicNamer) GetModelTopicInputs(modelName string) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, modelTopic, modelName, inputsSuffix)
}

func (tn *TopicNamer) GetModelTopicOutputs(modelName string) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, modelTopic, modelName, outputsSuffix)
}

func (tn *TopicNamer) GetPipelineTopicInputs(pipelineName string) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, pipelineTopic, pipelineName, inputsSuffix)
}

func (tn *TopicNamer) GetPipelineTopicOutputs(pipelineName string) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, pipelineTopic, pipelineName, outputsSuffix)
}

func (tn *TopicNamer) GetModelOrPipelineTopic(pipelineName string, stepReference string) string {
	stepName := strings.Split(stepReference, pipeline.StepNameSeperator)[0]
	if stepName == pipelineName {
		return fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, pipelineTopic, stepReference)
	} else {
		return fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, modelTopic, stepReference)
	}

}

func (tn *TopicNamer) GetFullyQualifiedTensorMap(pipelineName string, tin map[string]string) map[string]string {
	tout := make(map[string]string)
	for k, v := range tin {
		stepName := strings.Split(k, pipeline.StepNameSeperator)[0]
		var kout string
		if stepName == pipelineName {
			kout = fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, pipelineTopic, k)
		} else {
			kout = fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, modelTopic, k)
		}
		tout[kout] = v
	}
	return tout
}
