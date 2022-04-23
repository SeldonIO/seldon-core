package kafka

import (
	"fmt"
	"strings"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"
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
