package openai

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func getJsonBody(body []byte) (map[string]interface{}, error) {
	var jsonBody map[string]interface{}
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}

func getModelName(jsonBody map[string]interface{}) (string, error) {
	modelName, ok := jsonBody["model"].(string)
	if !ok {
		return "", nil
	}
	delete(jsonBody, "model")
	return modelName, nil
}

type Messages struct {
	Size       int
	Role       []string
	Content    []string
	Type       []string
	ToolCalls  []string
	ToolCallId []string
}

func NewMessages(size int) *Messages {
	return &Messages{
		Size:       size,
		Role:       make([]string, size),
		Content:    make([]string, size),
		Type:       make([]string, size),
		ToolCalls:  make([]string, size),
		ToolCallId: make([]string, size),
	}
}

func getMessageField(msgMap map[string]interface{}, field string) string {
	value, ok := msgMap[field]
	if !ok || value == nil {
		return ""
	}
	return value.(string)
}

func getContentAndType(msgMap map[string]interface{}) (string, string) {
	content, ok := msgMap["content"]
	if !ok {
		return "", ""
	}

	var contentType string
	var contentMessage string

	switch c := content.(type) {
	case string:
		contentType = "text"
		contentMessage = c

	case map[string]interface{}:
		t, ok := c["type"].(string)
		if !ok {
			return "", ""
		}
		contentType = t

		switch val := c[t].(type) {
		case string:
			contentMessage = val
		default:
			jsonContent, err := json.Marshal(val)
			if err != nil {
				contentMessage = ""
			} else {
				contentMessage = string(jsonContent)
			}
		}

	default:
		return "", ""
	}
	return contentMessage, contentType
}

func getMessages(jsonBody map[string]interface{}) (*Messages, error) {
	messagesList, ok := jsonBody["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("`messages` field not found or not an array")
	}

	delete(jsonBody, "messages")
	messages := NewMessages(len(messagesList))

	for i, message := range messagesList {
		msgMap, err := message.(map[string]interface{})
		if !err {
			return nil, fmt.Errorf("failed to parse message %d in OpenAI API request", i)
		}

		messages.Role[i] = getMessageField(msgMap, "role")
		messages.ToolCalls[i] = getMessageField(msgMap, "tool_calls")
		messages.ToolCallId[i] = getMessageField(msgMap, "tool_call_id")

		contentMessage, contentType := getContentAndType(msgMap)
		messages.Content[i] = contentMessage
		messages.Type[i] = contentType

	}
	return messages, nil
}

func getTools(jsonBody map[string]interface{}) (string, error) {
	tools, ok := jsonBody["tools"]
	delete(jsonBody, "tools")
	if !ok {
		return "", nil
	}

	data, err := json.Marshal(tools)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getLLMParameters(jsonBody map[string]interface{}) map[string]interface{} {
	llmParameters := make(map[string]interface{})
	for key, value := range jsonBody {
		if key == "model" || key == "messages" || key == "tools" {
			continue
		}
		llmParameters[key] = value
	}
	return llmParameters
}

func constructStringTensor(name string, shape []int, data []string) map[string]interface{} {
	return map[string]interface{}{
		"name":     name,
		"shape":    shape,
		"datatype": "BYTES",
		"data":     data,
	}
}

func constructSimpleInferenceRequestInputs(messages *Messages) []interface{} {
	return []interface{}{
		constructStringTensor("role", []int{1}, messages.Role),
		constructStringTensor("content", []int{1}, messages.Content),
		constructStringTensor("type", []int{1}, messages.Type),
		constructStringTensor("tool_calls", []int{1}, messages.ToolCalls),
		constructStringTensor("tool_call_id", []int{1}, messages.ToolCallId),
	}
}

func buildJSONContent(content []string) []string {
	jsonContent := make([]string, len(content))
	for i := range content {
		dataContent, err := json.Marshal([]string{content[i]})
		if err != nil {
			jsonContent[i] = ""
		} else {
			jsonContent[i] = string(dataContent)
		}
	}
	return jsonContent
}

func constructComplexInferenceRequestInputs(messages *Messages) []interface{} {
	return []interface{}{
		constructStringTensor("role", []int{1}, messages.Role),
		constructStringTensor("content", []int{1}, buildJSONContent(messages.Content)),
		constructStringTensor("type", []int{1}, buildJSONContent(messages.Type)),
		constructStringTensor("tool_calls", []int{1}, buildJSONContent(messages.ToolCalls)),
		constructStringTensor("tool_call_ids", []int{1}, buildJSONContent(messages.ToolCallId)),
	}
}

func constructInferenceRequest(messages *Messages, tools string, llmParams map[string]interface{}) map[string]interface{} {
	var inferenceRequestInputs []interface{}
	if messages.Size == 1 {
		inferenceRequestInputs = constructSimpleInferenceRequestInputs(messages)
	} else {
		inferenceRequestInputs = constructComplexInferenceRequestInputs(messages)
	}

	inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor("tools", []int{1}, []string{tools}))
	return map[string]interface{}{
		"inputs": inferenceRequestInputs,
		"parameters": map[string]interface{}{
			"llm_parameters": llmParams,
		},
	}
}

func ParserOpenAIAPI(body []byte, logger log.FieldLogger) (string, error) {
	// unmarshal the body to extract OpenAI API request
	jsonBody, err := getJsonBody(body)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API request body")
		return "", err
	}

	_, _ = getModelName(jsonBody)
	messages, err := getMessages(jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to parse messages in OpenAI API request")
		return "", err
	}
	tools, err := getTools(jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to parse tools in OpenAI API request")
		return "", err
	}
	llm_parameters := getLLMParameters(jsonBody)
	inferenceRequest := constructInferenceRequest(messages, tools, llm_parameters)

	data, err := json.Marshal(inferenceRequest)
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal OpenAI API request inputs")
		return "", err
	}
	return string(data), nil
}
