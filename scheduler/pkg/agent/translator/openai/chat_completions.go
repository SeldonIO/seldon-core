package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/translator"
)

type OpenAIChatCompletionsTranslator struct{}

func (t *OpenAIChatCompletionsTranslator) TranslateToOIP(req *http.Request, logger log.FieldLogger) (*http.Request, error) {
	// Read the request body
	body, err := translator.ReadRequestBody(req)
	if err != nil {
		logger.WithError(err).Error("Failed to read OpenAI API request body")
		return nil, err
	}

	jsonBody, err := translator.GetJsonBody(body)
	logger.Infof("Parsing OpenAI API request body %v", jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API request body")
		return nil, err
	}

	// Read model name. TODO: Check if the model name is in the request path
	_, _ = translator.GetModelName(jsonBody)

	// Parse messages from the request body
	messages, err := getMessages(jsonBody, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to parse messages in OpenAI API request")
		return nil, err
	}

	// Prepare tools and LLM parameters
	tools, _ := jsonBody["tools"].([]interface{})
	llm_parameters := getLLMParameters(jsonBody)

	// Construct the OIP formated input request
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

	// Create a new request with the translated body
	newBody := io.NopCloser(bytes.NewBuffer(data))
	newReq, err := http.NewRequest(req.Method, req.URL.String(), newBody)
	if err != nil {
		logger.WithError(err).Error("Failed to create new HTTP request for OpenAI API")
		return nil, err
	}
	newReq.Header = req.Header.Clone()

	// OpenAI API clinet adds `chat/completions` to the path, we need to remove it
	err = translator.TrimPathAfterInfer(newReq)
	if err != nil {
		logger.WithError(err).Error("Failed to trim path after infer in OpenAI API request")
		return nil, err
	}

	return newReq, nil
}

func (t *OpenAIChatCompletionsTranslator) TranslateFromOIP(res *http.Response, logger log.FieldLogger) (*http.Response, error) {
	// Decompress the response if needed - gzip
	var isGzipped bool
	var err error

	if isGzipped, err = translator.DecompressIfNeeded(res); err != nil {
		logger.WithError(err).Error("Failed to decompress OpenAI API response")
		return nil, err
	}

	// Read the response body
	body, err := translator.ReadResponseBody(res)
	if err != nil {
		logger.WithError(err).Error("Failed to read OpenAI API response body")
		return nil, err
	}

	jsonBody, err := translator.GetJsonBody(body)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API response body")
		return nil, err
	}

	// Parse the response body
	logger.Info("Parsing OpenAI API response body", jsonBody)
	outputs, ok := jsonBody["outputs"].([]interface{})
	if !ok {
		logger.Error("`outputs` field not found or not an array in OpenAI API response")
		return nil, fmt.Errorf("`outputs` field not found or not an array in OpenAI API response")
	}

	// Extract the output_all tensor form the inference response. This contains the full response
	// OpenAI API response - only works for OpenAI runtime, since we return the original OpenAI API response
	tensorName := "output_all"
	outputAll, err := translator.ExtractTensorByName(outputs, tensorName)
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

	// Create a new response with the translated body
	newBody := io.NopCloser(bytes.NewBuffer([]byte(content)))
	newRes := http.Response{
		StatusCode: res.StatusCode,
		Header:     res.Header.Clone(),
		Body:       newBody,
	}

	// compress the response body if needed
	if isGzipped {
		if err := translator.Compress(&newRes); err != nil {
			logger.WithError(err).Error("Failed to compress OpenAI API response")
			return nil, err
		}
	}
	return &newRes, nil
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

func getContentAndType(msgMap map[string]interface{}) ([]string, []string, error) {
	content, ok := msgMap["content"]
	if !ok || content == nil {
		return []string{""}, []string{"text"}, nil
	}

	switch c := content.(type) {
	case string:
		return []string{c}, []string{"text"}, nil
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

func getToolCalls(msgMap map[string]interface{}, logger log.FieldLogger) ([]string, error) {
	tcRaw, ok := msgMap["tool_calls"]
	if !ok {
		return nil, nil
	}

	if tcSlice, ok := tcRaw.([]interface{}); ok {
		toolCalls := make([]string, len(tcSlice))

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
		messages.Role[i], ok = msgMap["role"].(string)
		if !ok {
			return nil, fmt.Errorf("field 'role' not found in message %d", i)
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
		messages.ToolCallId[i] = msgMap["tool_call_id"]
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

func unwrapContentFromSlice(content []interface{}) ([]string, error) {
	switch c := content[0].(type) {
	case nil:
		return []string{}, nil
	case []string:
		return c, nil
	default:
		return marshalListContent(content)
	}
}

func prepareField(content []interface{}) ([]string, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("content is empty")
	}
	if len(content) == 1 {
		return unwrapContentFromSlice(content)
	}
	return marshalListContent(content)
}

func addFieldToInferenceRequestInputs(inferenceRequestInputs []interface{}, fieldName string, fieldContent []interface{}) ([]interface{}, error) {
	strContent, err := prepareField(fieldContent)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare field '%s' content: %v", fieldName, err)
	}
	if len(strContent) > 0 {
		inferenceRequestInputs = append(
			inferenceRequestInputs, translator.ConstructStringTensor(fieldName, strContent),
		)
	}
	return inferenceRequestInputs, nil
}

func constructInferenceRequestInputs(messages *Messages) ([]interface{}, error) {
	var inferenceRequestInputs []interface{}
	inferenceRequestInputs, err := addFieldToInferenceRequestInputs(
		inferenceRequestInputs, "role", messages.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'role' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, "content", messages.Content,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'content' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, "type", messages.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'type' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, "tool_calls", messages.ToolCalls,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add 'tool_calls' field to inference request inputs: %v", err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, "tool_call_id", messages.ToolCallId,
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
			translator.ConstructStringTensor("tools", strTools),
		)
	}

	return map[string]interface{}{
		"inputs": inferenceRequestInputs,
		"parameters": map[string]interface{}{
			"llm_parameters": llmParams,
		},
	}, nil
}
