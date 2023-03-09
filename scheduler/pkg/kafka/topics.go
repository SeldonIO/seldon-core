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
	"regexp"
	"strings"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

const (
	defaultSeldonTopicPrefix = "seldon"
	modelTopic               = "model"
	pipelineTopic            = "pipeline"
	errorsTopic              = "errors"
	inputsSuffix             = "inputs"
	outputsSuffix            = "outputs"
	errorsSuffix             = "errors"
	TopicErrorHeader         = "seldon-pipeline-errors"
	TopicSeparator           = "."
)

type TopicNamer struct {
	namespace   string
	topicPrefix string
}

var (
	isValidPrefix = regexp.MustCompile(`^[A-Za-z._-]+$`).MatchString
)

// Topics can only have a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash)
func NewTopicNamer(namespace string, topicPrefix string) (*TopicNamer, error) {
	if namespace == "" {
		namespace = "default"
	}
	if topicPrefix == "" {
		topicPrefix = defaultSeldonTopicPrefix
	}
	if !isValidPrefix(topicPrefix) {
		return nil, fmt.Errorf("The kafka prefix %s is invalid must only contain a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash) ", topicPrefix)
	}
	return &TopicNamer{
		namespace:   namespace,
		topicPrefix: topicPrefix,
	}, nil
}

func (tn *TopicNamer) GetModelErrorTopic() string {
	return strings.Join([]string{tn.topicPrefix, tn.namespace, errorsTopic, errorsSuffix}, TopicSeparator)
}

func (tn *TopicNamer) GetKafkaModelTopicRegex() string {
	return fmt.Sprintf("^%s%s.%s.*.%s", tn.topicPrefix, tn.namespace, modelTopic, inputsSuffix)
}

func (tn *TopicNamer) GetModelNameFromModelInputTopic(topic string) (string, error) {
	if !strings.HasPrefix(topic, tn.topicPrefix) {
		return "", fmt.Errorf("Topic does not have expectedTopic prefix %s. Topic = %s", tn.topicPrefix, topic)
	}
	parts := strings.Split(strings.TrimPrefix(topic, fmt.Sprintf("%s%s", tn.topicPrefix, TopicSeparator)), ".")
	if len(parts) < 4 {
		return "", fmt.Errorf("Wrong number of sections in topic %s. Was expecting 5 with separator '.'", topic)
	}
	if parts[0] != tn.namespace || parts[1] != modelTopic || parts[3] != inputsSuffix {
		return "", fmt.Errorf("Bad topic name %s needs to match %s", topic, tn.GetKafkaModelTopicRegex())
	}
	return parts[2], nil
}

func (tn *TopicNamer) GetModelTopicInputs(modelName string) string {
	return strings.Join([]string{tn.topicPrefix, tn.namespace, modelTopic, modelName, inputsSuffix}, TopicSeparator)
}

func (tn *TopicNamer) GetModelTopicOutputs(modelName string) string {
	return strings.Join([]string{tn.topicPrefix, tn.namespace, modelTopic, modelName, outputsSuffix}, TopicSeparator)
}

func (tn *TopicNamer) GetPipelineTopicInputs(pipelineName string) string {
	return strings.Join([]string{tn.topicPrefix, tn.namespace, pipelineTopic, pipelineName, inputsSuffix}, TopicSeparator)
}

func (tn *TopicNamer) GetPipelineTopicOutputs(pipelineName string) string {
	return strings.Join([]string{tn.topicPrefix, tn.namespace, pipelineTopic, pipelineName, outputsSuffix}, TopicSeparator)
}

func (tn *TopicNamer) GetModelOrPipelineTopicAndTensor(pipelineName string, stepReference string) (string, *string) {
	stepName := strings.Split(stepReference, pipeline.StepNameSeperator)[0]
	stepReference, tensor := tn.getTopicReferenceAndTensor(stepReference)
	if stepName == pipelineName {
		return strings.Join([]string{tn.topicPrefix, tn.namespace, pipelineTopic, stepReference}, TopicSeparator), tensor
	} else {
		return strings.Join([]string{tn.topicPrefix, tn.namespace, modelTopic, stepReference}, TopicSeparator), tensor
	}

}

func (tn *TopicNamer) GetFullyQualifiedTensorMap(pipelineName string, userTensorMap map[string]string) []*chainer.PipelineTensorMapping {
	var mappings []*chainer.PipelineTensorMapping
	for k, v := range userTensorMap {
		stepName := strings.Split(k, pipeline.StepNameSeperator)[0]
		var topicAndTensor string
		if stepName == pipelineName {
			topicAndTensor = strings.Join([]string{tn.topicPrefix, tn.namespace, pipelineTopic, k}, TopicSeparator)
		} else {
			topicAndTensor = strings.Join([]string{tn.topicPrefix, tn.namespace, modelTopic, k}, TopicSeparator)
		}
		mappings = append(mappings, &chainer.PipelineTensorMapping{
			PipelineName:   pipelineName,
			TopicAndTensor: topicAndTensor,
			TensorName:     v,
		})
	}
	return mappings
}

func (tn *TopicNamer) GetFullyQualifiedPipelineTensorMap(tin map[string]string) []*chainer.PipelineTensorMapping {
	var mappings []*chainer.PipelineTensorMapping
	for k, v := range tin {
		parts := strings.Split(k, pipeline.StepNameSeperator)
		var topicAndTensor string
		switch len(parts) {
		case 3: // A pipeline reference <pipelineName>.<inputs|outputs>.<tensorName>
			topicAndTensor = strings.Join([]string{tn.topicPrefix, tn.namespace, pipelineTopic, k}, TopicSeparator)
		case 5: // A fully qualified pipeline step tensor  <pipelineName>.step.<stepName>.<inputs|outputs>.<tensorName>
			// take value after <pipelineName>.step
			topicAndTensor = strings.Join([]string{tn.topicPrefix, tn.namespace, modelTopic, strings.Join(parts[2:], TopicSeparator)}, TopicSeparator)
		}
		mappings = append(mappings, &chainer.PipelineTensorMapping{
			PipelineName:   parts[0],
			TopicAndTensor: topicAndTensor,
			TensorName:     v,
		})
	}
	return mappings
}

// Always the initial part for pipeline inputs
// <pipelineName>. etc
func (tn *TopicNamer) GetPipelineNameFromInput(inputSpecifier string) string {
	return strings.Split(inputSpecifier, pipeline.StepNameSeperator)[0]
}

// if pipelineInput then return it: <pipelineName>.<inputs|outputs>
// If pipeline step then return stepName (which is a model reference) onwards: <pipelineName>.step.<stepName>.<inputs|outputs>.(tensorName)?
func (tn *TopicNamer) CreateStepReferenceFromPipelineInput(pipelineInputs string) string {
	parts := strings.Split(pipelineInputs, pipeline.StepNameSeperator)
	switch parts[1] {
	case pipeline.StepInputSpecifier, pipeline.StepOutputSpecifier:
		return pipelineInputs
	default:
		return strings.Join(parts[2:], pipeline.StepNameSeperator)
	}
}

// options are
// 1. step
// 2. step.(inputs|outputs)
// 3. step.(inputs|outputs).tensor
// Return the repference with tensor if there is one else just return reference
func (tn *TopicNamer) getTopicReferenceAndTensor(reference string) (string, *string) {
	parts := strings.Split(reference, TopicSeparator)
	numParts := len(parts)
	if numParts > 2 {
		tensor := parts[numParts-1]
		suffix := fmt.Sprintf("%s%s", TopicSeparator, tensor)
		return strings.TrimSuffix(reference, suffix), &tensor
	} else {
		return reference, nil
	}
}
