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
	Role       []interface{}
	Content    []interface{}
	Type       []interface{}
	ToolCalls  []interface{}
	ToolCallId []interface{}
}

func NewMessages(size int) *Messages {
	return &Messages{
		Size:       size,
		Role:       make([]interface{}, size),
		Content:    make([]interface{}, size),
		Type:       make([]interface{}, size),
		ToolCalls:  make([]interface{}, size),
		ToolCallId: make([]interface{}, size),
	}
}

func getMessageField(msgMap map[string]interface{}, field string) (interface{}, error) {
	value, ok := msgMap[field]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found in message map", field)
	}
	return value, nil
}

func getContentAndTypeFromMap(msgMap map[string]interface{}) (string, string, error) {
	contentType, ok := msgMap["type"].(string)
	if !ok {
		return "", "", fmt.Errorf("field 'type' not found or not a string in message map")
	}

	rawContentMessage, ok := msgMap[contentType]
	if !ok {
		return "", "", fmt.Errorf("field '%s' not found in message map", contentType)
	}

	var contentMessage string
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

func getContentAndType(msgMap map[string]interface{}) (interface{}, interface{}, error) {
	content, ok := msgMap["content"]
	if !ok || content == nil {
		return "", "text", nil
	}

	switch c := content.(type) {
	case string:
		return c, "text", nil
	case []interface{}:
		contentTypeArray := make([]string, len(c))
		contentMessageArray := make([]string, len(c))

		for i, item := range c {
			if itemMap, ok := item.(map[string]interface{}); ok {
				contentMessageI, contentTypeI, err := getContentAndTypeFromMap(itemMap)
				if err != nil {
					return nil, nil, err
				}
				contentMessageArray[i] = contentMessageI
				contentTypeArray[i] = contentTypeI
			}
		}
		return contentMessageArray, contentTypeArray, nil
	default:
		return nil, nil, fmt.Errorf("unsupported content type: %T", content)
	}
}

func getToolCalls(msgMap map[string]interface{}, logger log.FieldLogger) (interface{}, error) {
	tcRaw := msgMap["tool_calls"]
	logger.Infof("ToolCalls raw: %v", tcRaw)

	if tcRaw == nil {
		logger.Info("No tool calls found in message")
		return nil, nil // No tool calls present
	}

	logger.Infof("ToolCalls not nil: %v", tcRaw)
	if tcSlice, ok := tcRaw.([]interface{}); ok {
		toolCalls := make([]interface{}, len(tcSlice))
		for j, tc := range tcSlice {
			data, err := json.Marshal(tc)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tool call %d in message: %v", j, err)
			}
			toolCalls[j] = string(data)
		}
		return toolCalls, nil
	}

	return nil, fmt.Errorf("field 'tool_calls' is not a slice")
}

func getMessages(jsonBody map[string]interface{}, logger log.FieldLogger) (*Messages, error) {
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

		// Get role from the message map
		var err error
		messages.Role[i], err = getMessageField(msgMap, "role")
		if err != nil {
			return nil, fmt.Errorf("failed to get 'role' field in message %d: %v", i, err)
		}

		// Get content and type from the message map
		contentMessage, contentType, err := getContentAndType(msgMap)
		if err != nil {
			return nil, fmt.Errorf("failed to get content and type in message %d: %v", i, err)
		}
		messages.Content[i] = contentMessage
		messages.Type[i] = contentType

		// Get tool calls from the message map
		messages.ToolCalls[i], err = getToolCalls(msgMap, logger)
		logger.Infof("ToolCalls in message %d: %v", i, messages.ToolCalls[i])
		if err != nil {
			return nil, fmt.Errorf("failed to get tool calls in message %d: %v", i, err)
		}

		// Get tool call ID from the message map
		messages.ToolCallId[i], _ = getMessageField(msgMap, "tool_call_id")
	}
	return messages, nil
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

func wrapContentInSlice(content []interface{}) []interface{} {
	wrappedContent := make([]interface{}, len(content))
	for i, item := range content {
		if item == nil {
			wrappedContent[i] = nil
		} else {
			wrappedContent[i] = []interface{}{item}
		}
	}
	return wrappedContent
}

func marshalListContent(content []interface{}) ([]string, error) {
	jsonContent := make([]string, len(content))

	for i, item := range content {
		switch v := item.(type) {
		case nil:
			jsonContent[i] = ""
		case string:
			jsonContent[i] = v
		default:
			dataContent, err := json.Marshal(item)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal content item: %v", err)
			}
			jsonContent[i] = string(dataContent)
		}
	}
	return jsonContent, nil
}

func unwrapContentFromSlice(content []interface{}) []interface{} {
	switch c := content[0].(type) {
	case []interface{}:
		return c
	default:
		return content
	}
}

func prepareField(content []interface{}, wrapContent bool) ([]string, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("content is empty")
	}

	if len(content) == 1 {
		content = unwrapContentFromSlice(content)
		if content[0] == nil {
			return nil, nil
		}
		return marshalListContent(content)
	}

	if wrapContent {
		content = wrapContentInSlice(content)
	}
	return marshalListContent(content)
}

func addFieldToInferenceRequestInputs(
	inferenceRequestInputs []interface{},
	fieldName string,
	fieldContent []interface{},
	wrapContent bool,
) ([]interface{}, error) {
	strContent, err := prepareField(fieldContent, wrapContent)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare field '%s' content: %v", fieldName, err)
	}

	if strContent != nil {
		inferenceRequestInputs = append(inferenceRequestInputs, constructStringTensor(fieldName, strContent))
	}
	return inferenceRequestInputs, nil

}
func constructInferenceRequestInputs(messages *Messages) ([]interface{}, error) {
	var inferenceRequestInputs []interface{}
	inferenceRequestInputs, err := addFieldToInferenceRequestInputs(
		inferenceRequestInputs,
		"role",
		messages.Role,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'role' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs,
		"content",
		messages.Content,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'content' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs,
		"type",
		messages.Type,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'type' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs,
		"tool_calls",
		messages.ToolCalls,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'tool_calls' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs,
		"tool_call_id",
		messages.ToolCallId,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'tool_call_id' field to inference request inputs: %v", err)
	}
	return inferenceRequestInputs, nil
}

func constructInferenceRequest(messages *Messages, tools []interface{}, llmParams map[string]interface{}) (map[string]interface{}, error) {
	inferenceRequestInputs, err := constructInferenceRequestInputs(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to construct inference request inputs: %v", err)
	}

	if tools != nil {
		strTools, err := marshalListContent(tools)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tools content: %v", err)
		}
		inferenceRequestInputs = append(
			inferenceRequestInputs,
			constructStringTensor("tools", strTools),
		)
	}

	return map[string]interface{}{
		"inputs": inferenceRequestInputs,
		"parameters": map[string]interface{}{
			"llm_parameters": llmParams,
		},
	}, nil
}

func ParserOpenAIAPIRequest(body []byte, logger log.FieldLogger) ([]byte, error) {
	jsonBody, err := getJsonBody(body)
	logger.Infof("Parsing OpenAI API request body %v", jsonBody)

	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API request body")
		return nil, err
	}

	_, _ = getModelName(jsonBody)

	messages, err := getMessages(jsonBody, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to parse messages in OpenAI API request")
		return nil, err
	}

	tools, _ := jsonBody["tools"].([]interface{})
	llm_parameters := getLLMParameters(jsonBody)

	inferenceRequest, err := constructInferenceRequest(messages, tools, llm_parameters)
	if err != nil {
		logger.WithError(err).Error("Failed to construct inference request")
		return nil, err
	}

	data, err := json.Marshal(inferenceRequest)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal OpenAI API request inputs")
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
