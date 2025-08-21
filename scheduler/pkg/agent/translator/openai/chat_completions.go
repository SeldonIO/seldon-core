package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/translator"
)

type OpenAIChatCompletionsTranslator struct {
	translator.BaseTranslator
}

const (
	modelKey             = "model"
	messagesKey          = "messages"
	roleKey              = "role"
	contentKey           = "content"
	typeKey              = "type"
	toolsKey             = "tools"
	parallelToolCallsKey = "parallel_tool_calls"
	toolChoiceKey        = "tool_choice"
	toolCallsKey         = "tool_calls"
	toolCallIdKey        = "tool_call_id"
	inputsKey            = "inputs"
	parametersKey        = "parameters"
	llmParametersKey     = "llm_parameters"
)

func (t *OpenAIChatCompletionsTranslator) TranslateToOIP(req *http.Request) (*http.Request, error) {
	// Convert OpenAI API request to JSON
	jsonBody, err := translator.ConvertRequestToJsonBody(req)
	if err != nil {
		return nil, err
	}

	// Check if model name matches the one in the request path
	err = translator.CheckModelsMatch(jsonBody, req.URL.Path)
	if err != nil {
		return nil, err
	}

	// Parse messages from the request body
	messages, err := getMessages(jsonBody)
	if err != nil {
		return nil, err
	}

	// Prepare tools
	tools := getTools(jsonBody)
	parallelToolCalls := getParallelToolCalls(jsonBody)
	toolChoice := getToolChoice(jsonBody)

	// Prepare LLM parameters
	llm_parameters := getLLMParameters(jsonBody)

	// Construct the OIP formated input request
	inferenceRequest, err := constructChatCompletionInferenceRequest(messages, tools, parallelToolCalls, toolChoice, llm_parameters)
	if err != nil {
		return nil, err
	}

	// Construct new request
	return translator.ConvertInferenceRequestToHttpRequest(inferenceRequest, req)
}

func (t *OpenAIChatCompletionsTranslator) TranslateFromOIP(res *http.Response) (*http.Response, error) {
	httpRespones, err := t.BaseTranslator.TranslateFromOIP(res)
	if err == nil {
		return httpRespones, nil
	}

	if translator.IsServerSentEvent(res) {
		return t.translateStreamFromOIP(res)
	}

	return t.translateFromOIP(res)
}

func (t *OpenAIChatCompletionsTranslator) translateFromOIP(res *http.Response) (*http.Response, error) {
	jsonBody, isGzipped, err := translator.DecompressIfNeededAndConvertToJSON(res)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress and parse the response: %w", err)
	}

	id, ok := jsonBody["id"].(string)
	if !ok {
		return nil, fmt.Errorf("`id` field not found or not a string in the response")
	}

	modelName, ok := jsonBody["model_name"].(string)
	if !ok {
		return nil, fmt.Errorf("`model_name` field not found or not a string in the response")
	}

	outputs, ok := jsonBody[translator.OutputsKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("`%s` field not found or not an array in the response", translator.OutputsKey)
	}

	content, err := parseOuputChatCompletion(outputs, id, modelName, translator.IsServerSentEvent(res))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return translator.CreateResponseFromContent(
		content, res.StatusCode, res.Header, isGzipped,
	)
}

func (t *OpenAIChatCompletionsTranslator) translateStreamFromOIP(res *http.Response) (*http.Response, error) {
	pr, pw := io.Pipe()

	// declare the scanner and override the default split function
	scanner := bufio.NewScanner(res.Body)
	scanner.Split(translator.SplitSSE)

	// read first line which faild to parse
	firstLine := res.Header.Get(translator.FirstLineKey)
	res.Header.Del(translator.FirstLineKey)

	translated, err := translateLocalLine(firstLine)
	if err != nil {
		pw.CloseWithError(err)
		return nil, fmt.Errorf("failed to translate first line: %w", err)
	}

	go translator.ScanAndWriteSSE(res, &translated, scanner, pw, translateLocalLine)

	// Return single streaming response
	return &http.Response{
		StatusCode: res.StatusCode,
		Header:     res.Header.Clone(),
		Body:       pr,
	}, nil
}

func translateLocalLine(line string) (string, error) {
	line = strings.TrimPrefix(line, translator.SSEPrefix)
	jsonLine, err := translator.GetJsonBody([]byte(line))
	if err != nil {
		return "", fmt.Errorf("failed to parse SSE line: %w", err)
	}

	id, ok := jsonLine["id"].(string)
	if !ok {
		return "", fmt.Errorf("`id` field not found or not a string in the response")
	}

	modelName, ok := jsonLine["model_name"].(string)
	if !ok {
		return "", fmt.Errorf("`model_name` field not found or not a string in the response")
	}

	outputs, ok := jsonLine[translator.OutputsKey].([]interface{})
	if !ok {
		return "", fmt.Errorf("`%s` field not found or not an array in the response", translator.OutputsKey)
	}

	content, err := parseOuputChatCompletion(outputs, id, modelName, true)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return fmt.Sprintf("%s%s%s", translator.SSEPrefix, string(content), translator.SSESuffix), nil

}

func getTools(jsonBody map[string]interface{}) []interface{} {
	tools, _ := jsonBody[toolsKey].([]interface{})
	return tools
}

func getParallelToolCalls(jsonBody map[string]interface{}) []interface{} {
	parallelToolCalls, ok := jsonBody[parallelToolCallsKey]
	if !ok {
		return []interface{}{}
	}
	return []interface{}{parallelToolCalls}
}

func getToolChoice(jsonBody map[string]interface{}) []interface{} {
	toolChoice, ok := jsonBody[toolChoiceKey]
	if !ok {
		return []interface{}{}
	}
	return []interface{}{toolChoice}
}

func parseOuputChatCompletion(outputs []interface{}, id string, modelName string, isStream bool) (string, error) {
	role, err := translator.ExtractTensorContentFromResponse(outputs, roleKey)
	if err != nil {
		return "", err
	}

	content, err := translator.ExtractTensorContentFromResponse(outputs, contentKey)
	if err != nil {
		return "", err
	}

	// construct the OpenAI API response
	var response map[string]interface{}
	if isStream {
		response = map[string]interface{}{
			"id":      id,
			"model":   modelName,
			"created": 0,
			"object":  "chat.completion.chunk",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"delta": map[string]interface{}{
						"role":    role,
						"content": content,
					},
				},
			},
		}
	} else {
		response = map[string]interface{}{
			"id":      id,
			"model":   modelName,
			"created": 0,
			"object":  "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"content": content,
						"role":    role,
					},
				},
			},
		}
	}

	// convert the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OpenAI API response: %w", err)
	}
	return string(jsonResponse), nil
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
	contentType, ok := msgMap[typeKey].(string)
	if !ok {
		return "", "", fmt.Errorf("field '%s' not found or not a string in message map", typeKey)
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
	content, ok := msgMap[contentKey]
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

func getToolCalls(msgMap map[string]interface{}) ([]string, error) {
	tcRaw, ok := msgMap[toolCallsKey]
	if !ok {
		return []string{}, nil
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

	return nil, fmt.Errorf("field '%s' is not a slice", toolCallsKey)
}

func getMessages(jsonBody map[string]interface{}) (*Messages, error) {
	messagesList, ok := jsonBody[messagesKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("`%s` field not found or not an array", messagesKey)
	}

	delete(jsonBody, messagesKey)
	messages := NewMessages(len(messagesList))

	for i, message := range messagesList {
		msgMap, boolErr := message.(map[string]interface{})
		if !boolErr {
			return nil, fmt.Errorf("failed to parse message %d in OpenAI API request", i)
		}

		// Get role from the message map
		var err error
		messages.Role[i], ok = msgMap[roleKey].(string)
		if !ok {
			return nil, fmt.Errorf("field '%s' not found in message %d", roleKey, i)
		}

		// Get content and type from the message map
		contentMessage, contentType, err := getContentAndType(msgMap)
		if err != nil {
			return nil, fmt.Errorf("failed to get content and type in message %d: %v", i, err)
		}
		messages.Content[i] = contentMessage
		messages.Type[i] = contentType

		// Get tool calls from the message map
		messages.ToolCalls[i], err = getToolCalls(msgMap)
		if err != nil {
			return nil, fmt.Errorf("failed to get tool calls in message %d: %v", i, err)
		}

		// Get tool call ID from the message map
		messages.ToolCallId[i], _ = msgMap[toolCallIdKey].(string)
	}
	return messages, nil
}

func getLLMParameters(jsonBody map[string]interface{}) map[string]interface{} {
	llmParameters := make(map[string]interface{})
	skipKeys := []string{
		modelKey, messagesKey, toolsKey, parallelToolCallsKey, toolChoiceKey,
	}
	for key, value := range jsonBody {
		if translator.Contains(skipKeys, key) {
			continue
		}
		llmParameters[key] = value
	}
	return llmParameters
}

func marshalListContent(content []interface{}) ([]string, error) {
	jsonContent := make([]string, len(content))

	for i, item := range content {
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Slice && val.Len() == 0 {
			jsonContent[i] = ""
			continue
		}

		switch v := item.(type) {
		case string:
			jsonContent[i] = v
		default:
			dataContent, err := json.Marshal(v)
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
	case string:
		if len(c) == 0 {
			return []string{}, nil
		}
		return []string{c}, nil
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
	if len(strContent) > 0 && !translator.IsEmptySlice(strContent) {
		inferenceRequestInputs = append(
			inferenceRequestInputs, translator.ConstructStringTensor(fieldName, strContent),
		)
	}
	return inferenceRequestInputs, nil
}

func constructInferenceRequestInputs(messages *Messages) ([]interface{}, error) {
	var inferenceRequestInputs []interface{}
	inferenceRequestInputs, err := addFieldToInferenceRequestInputs(
		inferenceRequestInputs, roleKey, messages.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add '%s' field to inference request inputs: %v", roleKey, err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, contentKey, messages.Content,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add '%s' field to inference request inputs: %v", contentKey, err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, typeKey, messages.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add '%s' field to inference request inputs: %v", typeKey, err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, toolCallsKey, messages.ToolCalls,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add '%s' field to inference request inputs: %v", toolCallsKey, err)
	}

	inferenceRequestInputs, err = addFieldToInferenceRequestInputs(
		inferenceRequestInputs, toolCallIdKey, messages.ToolCallId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add '%s' field to inference request inputs: %v", toolCallIdKey, err)
	}
	return inferenceRequestInputs, nil
}

func constructChatCompletionInferenceRequest(messages *Messages, tools []interface{}, parallelToolCalls []interface{}, toolChoice []interface{}, llmParams map[string]interface{}) (map[string]interface{}, error) {
	inferenceRequestInputs, err := constructInferenceRequestInputs(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to construct inference request inputs: %v", err)
	}

	if len(tools) > 0 {
		strTools, err := marshalListContent(tools)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %s content: %v", toolsKey, err)
		}
		inferenceRequestInputs = append(
			inferenceRequestInputs,
			translator.ConstructStringTensor(toolsKey, strTools),
		)

		// Add parallel_tool_calls if present
		if len(parallelToolCalls) > 0 {
			strParallelToolCalls, err := marshalListContent(parallelToolCalls)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal %s content: %v", parallelToolCallsKey, err)
			}
			inferenceRequestInputs = append(
				inferenceRequestInputs,
				translator.ConstructStringTensor(parallelToolCallsKey, strParallelToolCalls),
			)
		}

		// Add tool_choice if present
		if len(toolChoice) > 0 {
			strToolChoice, err := marshalListContent(toolChoice)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal %s content: %v", toolChoiceKey, err)
			}
			inferenceRequestInputs = append(
				inferenceRequestInputs,
				translator.ConstructStringTensor(toolChoiceKey, strToolChoice),
			)
		}
	}

	return map[string]interface{}{
		inputsKey: inferenceRequestInputs,
		parametersKey: map[string]interface{}{
			llmParametersKey: llmParams,
		},
	}, nil
}
