package kafka

import (
	"fmt"
)

const (
	seldonTopicPrefix = "seldon"
	modelTopic        = "model"
	pipelineTopic     = "pipeline"
	errorsTopic       = "errors"
	inputsSuffix      = "inputs"
	outputsSuffix     = "outputs"
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
	return fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, errorsTopic, outputsSuffix)
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

func (tn *TopicNamer) GetModelTopic(modelStream string) string {
	return fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, modelTopic, modelStream)
}

func (tn *TopicNamer) GetFullyQualifiedTensorMap(tin map[string]string) map[string]string {
	tout := make(map[string]string)
	for k, v := range tin {
		kout := fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, tn.namespace, modelTopic, k)
		tout[kout] = v
	}
	return tout
}
