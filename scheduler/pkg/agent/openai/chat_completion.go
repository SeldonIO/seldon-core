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

func getMessageField(msgMap map[string]interface{}, field string) (string, error) {
	value, ok := msgMap[field]
	if !ok || value == nil {
		return "", fmt.Errorf("field '%s' not found in message map", field)
	}
	return value.(string), nil
}

func getToolCalls(msgMap map[string]interface{}) (string, error) {
	toolCalls, ok := msgMap["tool_calls"].([]interface{})
	if !ok {
		return "", nil
	}

	toolCallsData, err := json.Marshal(toolCalls)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool calls: %v", err)
	}

	return string(toolCallsData), nil
}

func getContentAndTypeFromMap(msgMap map[string]interface{}) (string, string, error) {
	contentType, ok := msgMap["type"].(string)
	if !ok {
		return "", "", fmt.Errorf("field 'type' not found or not a string in message map")
	}

	var contentMessage string

	rawContentMessage, ok := msgMap[contentType]
	if !ok {
		return "", "", fmt.Errorf("field '%s' not found in message map", contentType)
	}

	switch c := rawContentMessage.(type) {
	case string:
		contentMessage = c
	case map[string]interface{}:
		data, err := json.Marshal(c)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal content message map")
		}
		contentMessage = string(data)
	default:
		return "", "", fmt.Errorf("unsupported content type: %T", rawContentMessage)
	}

	return contentMessage, contentType, nil
}

func getContentAndType(msgMap map[string]interface{}) (string, string, error) {
	content, ok := msgMap["content"]
	if !ok {
		return "", "", fmt.Errorf("field 'content' not found in message map")
	}

	var contentType string
	var contentMessage string

	switch c := content.(type) {
	case string:
		contentType = "text"
		contentMessage = c

	case []interface{}:
		contentTypeArray := make([]string, len(c))
		contentMessageArray := make([]string, len(c))

		for _, item := range c {
			if itemMap, ok := item.(map[string]interface{}); ok {
				contentMessageI, contentTypeI, err := getContentAndTypeFromMap(itemMap)
				if err != nil {
					return "", "", err
				}
				contentMessageArray = append(contentMessageArray, contentMessageI)
				contentTypeArray = append(contentTypeArray, contentTypeI)
			}
		}

		bytesContentMessage, err := json.Marshal(contentMessageArray)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal content messages: %v", err)
		}

		bytestContentType, err := json.Marshal(contentTypeArray)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal content types: %v", err)
		}

		contentType = string(bytestContentType)
		contentMessage = string(bytesContentMessage)
	default:
		return "", "", fmt.Errorf("unsupported content type: %T", content)
	}
	return contentMessage, contentType, nil
}

func getMessages(jsonBody map[string]interface{}) (*Messages, error) {
	messagesList, ok := jsonBody["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("`messages` field not found or not an array")
	}

	delete(jsonBody, "messages")
	messages := NewMessages(len(messagesList))

	for i, message := range messagesList {
		msgMap, boolErr := message.(map[string]interface{})
		if !boolErr {
			return nil, fmt.Errorf("failed to parse message %d in OpenAI API request", i)
		}

		var err error
		messages.Role[i], err = getMessageField(msgMap, "role")
		if err != nil {
			return nil, fmt.Errorf("failed to get 'role' field in message %d: %v", i, err)
		}

		contentMessage, contentType, err := getContentAndType(msgMap)
		if err != nil {
			return nil, fmt.Errorf("failed to get content and type in message %d: %v", i, err)
		}

		messages.Content[i] = contentMessage
		messages.Type[i] = contentType

		// Optional fields, so we ignore errors
		messages.ToolCalls[i], _ = getToolCalls(msgMap)
		messages.ToolCallId[i], _ = getMessageField(msgMap, "tool_call_id")
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

func constructStringTensor(name string, data []string) map[string]interface{} {
	return map[string]interface{}{
		"name":     name,
		"shape":    []int{len(data)},
		"datatype": "BYTES",
		"data":     data,
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

func constructSimpleInferenceRequestInputs(messages *Messages) []interface{} {
	inferenceRequestInputs := make([]interface{}, 0, 5)
	inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor("role", messages.Role))
	inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor("content", messages.Content))
	inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor("type", messages.Type))
	if messages.ToolCalls[0] != "" {
		inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor("tool_calls", messages.ToolCalls))
	}
	if messages.ToolCallId[0] != "" {
		inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor("tool_call_id", messages.ToolCallId))
	}
	return inferenceRequestInputs
}

func constructComplexInferenceRequestInputs(messages *Messages) []interface{} {
	return []interface{}{
		constructStringTensor("role", messages.Role),
		constructStringTensor("content", buildJSONContent(messages.Content)),
		constructStringTensor("type", buildJSONContent(messages.Type)),
		constructStringTensor("tool_calls", messages.ToolCalls),
		constructStringTensor("tool_call_ids", messages.ToolCallId),
	}
}

func constructInferenceRequest(messages *Messages, tools string, llmParams map[string]interface{}) map[string]interface{} {
	var inferenceRequestInputs []interface{}
	if messages.Size == 1 {
		inferenceRequestInputs = constructSimpleInferenceRequestInputs(messages)
	} else {
		inferenceRequestInputs = constructComplexInferenceRequestInputs(messages)
	}

	// inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor("tools", []string{tools}))
	return map[string]interface{}{
		"inputs": inferenceRequestInputs,
		"parameters": map[string]interface{}{
			"llm_parameters": llmParams,
		},
	}
}

func ParserOpenAIAPIRequest(body []byte, logger log.FieldLogger) ([]byte, error) {
	jsonBody, err := getJsonBody(body)
	logger.Infof("Parsing OpenAI API request body %v", jsonBody)

	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API request body")
		return nil, err
	}

	_, _ = getModelName(jsonBody)

	messages, err := getMessages(jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to parse messages in OpenAI API request")
		return nil, err
	}

	tools, err := getTools(jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to parse tools in OpenAI API request")
		return nil, err
	}

	llm_parameters := getLLMParameters(jsonBody)

	inferenceRequest := constructInferenceRequest(messages, tools, llm_parameters)
	data, err := json.Marshal(inferenceRequest)
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal OpenAI API request inputs")
		return nil, err
	}
	return data, nil
}

func extractTensorByName(outputs []interface{}, name string) (map[string]interface{}, error) {
	for i, output := range outputs {
		outputMap, ok := output.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse output tensor %d", i)
		}

		if outputMap["name"] == name {
			return outputMap, nil
		}
	}
	return nil, fmt.Errorf("output tensor with name %s not found", name)
}

func ParseOpenAIAPIResponse(body []byte, logger log.FieldLogger) ([]byte, error) {
	jsonBody, err := getJsonBody(body)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API response body")
		return nil, err
	}

	logger.Info("Parsing OpenAI API response body", jsonBody)
	outputs, ok := jsonBody["outputs"].([]interface{})
	if !ok {
		logger.Error("`outputs` field not found or not an array in OpenAI API response")
		return nil, fmt.Errorf("`outputs` field not found or not an array in OpenAI API response")
	}

	tensorName := "output_all"
	outputAll, err := extractTensorByName(outputs, tensorName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to extract '%s' tensor from OpenAI API response", tensorName)
		return nil, err
	}

	data, ok := outputAll["data"].([]interface{})
	if !ok {
		logger.Errorf("`data` field not found or not an array of strings in output tensor %s", tensorName)
		return nil, fmt.Errorf("`data` field not found or not an array of strings in output tensor %s", tensorName)
	}

	content, ok := data[0].(string)
	if !ok {
		logger.Errorf("`data` field in output tensor %s is not a byte array", tensorName)
		return nil, fmt.Errorf("`data` field in output tensor %s is not a byte array", tensorName)
	}
	return []byte(content), nil
}

func GetConstantResponse() string {
	return `{"choices":[{"message":{"role":"assistant","content":"This is a constant response from the OpenAI API parser."},"finish_reason":"stop","index":0}],"created":1700000000,"id":"chatcmpl-1234567890","model":"gpt-3.5-turbo","object":"chat.completion","usage":{"prompt_tokens":10,"completion_tokens":10,"total_tokens":20}}`
}
