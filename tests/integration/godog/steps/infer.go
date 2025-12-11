/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	"google.golang.org/grpc/metadata"
)

func (i *inference) doHTTPModelInferenceRequest(ctx context.Context, modelName, body string) error {
	url := fmt.Sprintf(
		"%s://%s:%d/v2/models/%s/infer",
		httpScheme(i.ssl),
		i.host,
		i.httpPort,
		modelName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("could not create http request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", "seldon-mesh.inference.seldon")
	req.Header.Add("Seldon-model", modelName)

	resp, err := i.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not send http request: %w", err)
	}

	i.lastHTTPResponse = resp
	return nil
}

// Used from steps that pass an explicit payload (DocString)
func (i *inference) sendHTTPModelInferenceRequest(ctx context.Context, model string, payload *godog.DocString) error {
	return i.doHTTPModelInferenceRequest(ctx, model, payload.Content)
}

// Used from steps that work from a *Model and testModels table
func (i *inference) sendHTTPModelInferenceRequestFromModel(ctx context.Context, m *Model) error {
	testModel, ok := testModels[m.modelType]
	if !ok {
		return fmt.Errorf("could not find test model %s", m.model.Name)
	}

	return i.doHTTPModelInferenceRequest(ctx, m.modelName, testModel.ValidInferenceRequest)
}

func httpScheme(useSSL bool) string {
	if useSSL {
		return "https"
	}
	return "http"
}

func (i *inference) sendGRPCModelInferenceRequest(ctx context.Context, model string, payload *godog.DocString) error {
	return i.doGRPCModelInferenceRequest(ctx, model, payload.Content)
}

func (i *inference) sendGRPCModelInferenceRequestFromModel(ctx context.Context, m *Model) error {
	testModel, ok := testModels[m.modelType]
	if !ok {
		return fmt.Errorf("could not find test model %s", m.model.Name)
	}
	return i.doGRPCModelInferenceRequest(ctx, m.modelName, testModel.ValidInferenceRequest)
}

func (i *inference) doGRPCModelInferenceRequest(
	ctx context.Context,
	model string,
	payload string,
) error {
	// Unmarshal into a value, then take its address when calling gRPC.
	var req v2_dataplane.ModelInferRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return fmt.Errorf("could not unmarshal gRPC json payload: %w", err)
	}
	req.ModelName = model

	// Attach metadata to the *existing* context, don’t discard it.
	md := metadata.Pairs("seldon-model", model)
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := i.grpc.ModelInfer(ctx, &req)

	// Record both resp and err so later steps can assert on them.
	i.lastGRPCResponse.response = resp
	i.lastGRPCResponse.err = err

	// Important: return nil so that the step itself doesn't fail.
	// The following "Then ..." step will assert on i.lastGRPCResponse.err.
	return nil
}

func withTimeoutCtx(timeout string, callback func(ctx context.Context) error) error {
	ctx, cancel, err := timeoutToContext(timeout)
	if err != nil {
		return err
	}
	defer cancel()
	return callback(ctx)
}

func timeoutToContext(timeout string) (context.Context, context.CancelFunc, error) {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid timeout %s: %w", timeout, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), d)
	return ctx, cancel, nil
}

func isSubset(needle, hay any) bool {
	nObj, nOK := needle.(map[string]any)
	hObj, hOK := hay.(map[string]any)
	if nOK && hOK {
		for k, nv := range nObj {
			hv, exists := hObj[k]
			if !exists || !isSubset(nv, hv) {
				return false
			}
		}
		return true
	}

	return reflect.DeepEqual(needle, hay)
}

func containsSubset(needle, hay any) bool {
	if isSubset(needle, hay) {
		return true
	}
	switch h := hay.(type) {
	case map[string]any:
		for _, v := range h {
			if containsSubset(needle, v) {
				return true
			}
		}
	case []any:
		for _, v := range h {
			if containsSubset(needle, v) {
				return true
			}
		}
	}
	return false
}

func jsonContainsObjectSubset(jsonStr, needleStr string) (bool, error) {
	var hay, needle any
	if err := json.Unmarshal([]byte(jsonStr), &hay); err != nil {
		return false, fmt.Errorf("could not unmarshal hay json %s: %w", jsonStr, err)
	}
	if err := json.Unmarshal([]byte(needleStr), &needle); err != nil {
		return false, fmt.Errorf("could not unmarshal needle json %s: %w", needleStr, err)
	}
	return containsSubset(needle, hay), nil
}

func (i *inference) gRPCRespContainsError(err string) error {
	if i.lastGRPCResponse.err == nil {
		return errors.New("no gRPC response error found")
	}

	if strings.Contains(i.lastGRPCResponse.err.Error(), err) {
		return nil
	}

	return fmt.Errorf("error %s does not contain %s", i.lastGRPCResponse.err.Error(), err)
}

func (i *inference) gRPCRespCheckBodyContainsJSON(expectJSON *godog.DocString) error {
	if i.lastGRPCResponse.response == nil {
		return errors.New("no gRPC response found")
	}

	gotJson, err := json.Marshal(i.lastGRPCResponse.response)
	if err != nil {
		return fmt.Errorf("could not marshal gRPC json: %w", err)
	}

	ok, err := jsonContainsObjectSubset(string(gotJson), expectJSON.Content)
	if err != nil {
		return fmt.Errorf("could not check if json contains object: %w", err)
	}

	if !ok {
		return fmt.Errorf("%s does not contain %s", string(gotJson), expectJSON)
	}

	return nil
}

func (i *inference) httpRespCheckBodyContainsJSON(expectJSON *godog.DocString) error {
	if i.lastHTTPResponse == nil {
		return errors.New("no http response found")
	}

	body, err := io.ReadAll(i.lastHTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	ok, err := jsonContainsObjectSubset(string(body), expectJSON.Content)
	if err != nil {
		return fmt.Errorf("could not check if json contains object: %w", err)
	}

	if !ok {
		return fmt.Errorf("%s does not contain %s", string(body), expectJSON)
	}

	return nil
}

func (i *inference) httpRespCheckStatus(status int) error {
	if i.lastHTTPResponse == nil {
		return errors.New("no http response found")
	}
	if status != i.lastHTTPResponse.StatusCode {
		body, err := io.ReadAll(i.lastHTTPResponse.Body)
		if err != nil {
			return fmt.Errorf("expected http response status code %d, got %d", status, i.lastHTTPResponse.StatusCode)
		}
		return fmt.Errorf("expected http response status code %d, got %d with body: %s", status, i.lastHTTPResponse.StatusCode, body)

	}
	return nil
}
