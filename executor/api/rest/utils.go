package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/seldonio/seldon-core/executor/api/payload"
)

const (
	// OpenAPI values
	openapiFilePath     = "./openapi/seldon.json"
	openapiPredPath     = "/seldon/{namespace}/{deployment}/api/v1.0/predictions"
	openapiFeedbackPath = "/seldon/{namespace}/{deployment}/api/v1.0/feedback"
)

func CombineSeldonMessagesToJson(msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	// Extract into string array checking the data is JSON
	strData := make([]string, len(msgs))
	for i, sm := range msgs {
		bytes, err := sm.GetBytes()
		if err != nil {
			return nil, err
		}
		if !isJSON(bytes) {
			return nil, fmt.Errorf("Data is not JSON")
		} else {
			strData[i] = string(sm.GetPayload().([]byte))
		}
	}
	// Create JSON list of messages
	joined := strings.Join(strData, ",")
	jStr := "[" + joined + "]"
	return &payload.BytesPayload{Msg: []byte(jStr), ContentType: msgs[0].GetContentType()}, nil
}

func ExtractSeldonMessagesFromJson(msg payload.SeldonPayload) ([]payload.SeldonPayload, error) {
	bytes, err := msg.GetBytes()
	if err != nil {
		return nil, err
	}
	var v []interface{}
	sms := make([]payload.SeldonPayload, len(v))
	err = json.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}
	for _, m := range v {
		mBytes, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		sms = append(sms, &payload.BytesPayload{Msg: mBytes})
	}
	return sms, nil
}

func embedSeldonDeploymentValuesInJson(namespace string, sdepName string, openapiInterface *interface{}) error {
	jsonFormatError := errors.New("Incorrect format for OpenAPI JSON file")

	replacer := strings.NewReplacer(
		"{namespace}", namespace,
		"{deployment}", sdepName)

	// Ensure json is a map before performing actions
	if openapiJson, ok := (*openapiInterface).(map[string]interface{}); ok {
		// Remove the servers element to ensure it uses the same URL
		delete(openapiJson, "servers")

		// Get the paths key value to remove the parameters from each of the URLs
		if pathsJson, ok := openapiJson["paths"].(map[string]interface{}); ok {
			// Delete the parameters field from the predictions path
			if openapiPredPathJson, ok := pathsJson[openapiPredPath].(map[string]interface{}); ok {
				if openapiPredPathPostJson, ok := openapiPredPathJson["post"].(map[string]interface{}); ok {
					delete(openapiPredPathPostJson, "parameters")
				} else {
					return jsonFormatError
				}
			} else {
				return jsonFormatError
			}

			// Rename the predictions path to use the namespace and deploymentName instead of placeholder values
			openapiPredPathReplaced := replacer.Replace(openapiPredPath)
			pathsJson[openapiPredPathReplaced] = pathsJson[openapiPredPath]
			delete(pathsJson, openapiPredPath)

			// Delete the parameters field from the feedback path
			if openapiFeedbackPathJson, ok := pathsJson[openapiFeedbackPath].(map[string]interface{}); ok {
				if openapiFeedbackPathPostJson, ok := openapiFeedbackPathJson["post"].(map[string]interface{}); ok {
					delete(openapiFeedbackPathPostJson, "parameters")
				} else {
					return jsonFormatError
				}
			} else {
				return jsonFormatError
			}

			// Rename the predictions path to use the namespace and deploymentName instead of placeholder values
			openapiFeedbackPathReplaced := replacer.Replace(openapiFeedbackPath)
			pathsJson[openapiFeedbackPathReplaced] = pathsJson[openapiFeedbackPath]
			delete(pathsJson, openapiFeedbackPath)
		} else {
			return jsonFormatError
		}

	} else {
		return jsonFormatError
	}

	return nil
}

func EmbedSeldonDeploymentValuesInSwaggerFile(namespace string, sdepName string) error {
	openapiInputBytes, err := ioutil.ReadFile(openapiFilePath)
	if err != nil {
		return err
	}
	var openapiInterface interface{}
	if err := json.Unmarshal(openapiInputBytes, &openapiInterface); err != nil {
		return err
	}

	if err := embedSeldonDeploymentValuesInJson(namespace, sdepName, &openapiInterface); err != nil {
		return err
	}

	// We use MarshalIndent so that the output can be humanly visible and indented
	openapiOutputBytes, err := json.MarshalIndent(openapiInterface, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(openapiFilePath, openapiOutputBytes, 0644); err != nil {
		return err
	}

	return nil
}

func isJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
