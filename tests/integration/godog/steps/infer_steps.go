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
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type inference struct {
	ssl              bool
	host             string
	http             *http.Client
	grpc             v2_dataplane.GRPCInferenceServiceClient
	httpPort         uint
	lastHTTPResponse *http.Response
	lastGRPCResponse lastGRPCResponse
	log              logrus.FieldLogger
}

// todo: this is to avoid 503s since the route at times isn't ready when the model is ready
// todo: this might be related to issue: https://seldonio.atlassian.net/browse/INFRA-1576?search_id=4def05ca-ec64-436e-b824-49e39f1c94c4
// todo: this issue relates to how we should route inference request via IP instead of DNS
const inferenceRequestDelay = 200 * time.Millisecond

func LoadInferenceSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^(?:I )send HTTP inference request with timeout "([^"]+)" to (model|pipeline) "([^"]+)" with payload:$`, func(timeout, kind, resourceName string, payload *godog.DocString) error {
		time.Sleep(inferenceRequestDelay)
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch kind {
			case "model":
				return w.infer.doHTTPModelInferenceRequest(ctx, resourceName, payload.Content)
			case "pipeline":
				return w.infer.doHTTPPipelineInferenceRequest(ctx, resourceName, payload.Content)
			default:
				return fmt.Errorf("unknown target type: %s", kind)
			}
		})
	})
	scenario.Step(`^(?:I )send gRPC inference request with timeout "([^"]+)" to (model|pipeline) "([^"]+)" with payload:$`, func(timeout, kind, resourceName string, payload *godog.DocString) error {
		time.Sleep(inferenceRequestDelay)
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch kind {
			case "model":
				return w.infer.doGRPCInferenceRequest(ctx, resourceName, payload.Content)
			case "pipeline":
				return w.infer.doGRPCInferenceRequest(ctx, fmt.Sprintf("%s.pipeline", resourceName), payload.Content)
			default:
				return fmt.Errorf("unknown target type: %s", kind)
			}
		})
	})
	scenario.Step(`^(?:I )send a valid gRPC inference request with timeout "([^"]+)"`, func(timeout string) error {
		time.Sleep(inferenceRequestDelay)
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.infer.sendGRPCModelInferenceRequestFromModel(ctx, w.currentModel)
		})
	})
	scenario.Step(`^(?:I )send a valid HTTP inference request with timeout "([^"]+)"`, func(timeout string) error {
		time.Sleep(inferenceRequestDelay)
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.infer.sendHTTPModelInferenceRequestFromModel(ctx, w.currentModel)
		})
	})

	scenario.Step(`^expect http response status code "([^"]*)"$`, w.infer.httpRespCheckStatus)
	scenario.Step(`^expect http response body to contain JSON:$`, w.infer.httpRespCheckBodyContainsJSON)
	scenario.Step(`^expect gRPC response body to contain JSON:$`, w.infer.gRPCRespCheckBodyContainsJSON)
	scenario.Step(`^expect gRPC response error to contain "([^"]+)"`, w.infer.gRPCRespContainsError)
	scenario.Step(`^expect gRPC response to not return an error$`, w.infer.gRPCRespContainsNoError)
	scenario.Step(`^expect http response body to contain valid JSON$`, func() error {
		testModel, ok := testModels[w.currentModel.modelType]
		if !ok {
			return fmt.Errorf("model %s not found", w.currentModel.modelType)
		}
		return w.infer.doHttpRespCheckBodyContainsJSON(testModel.ValidJSONResponse)
	})
}

func (i *inference) doHTTPInferenceRequest(ctx context.Context, resourceName, headerName, body string) error {
	url := fmt.Sprintf(
		"%s://%s:%d/v2/models/%s/infer",
		httpScheme(i.ssl),
		i.host,
		i.httpPort,
		resourceName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("could not create http request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", "seldon-mesh.inference.seldon")
	req.Header.Add("Seldon-model", headerName)

	resp, err := i.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not send http request: %w", err)
	}

	i.lastHTTPResponse = resp
	return nil
}

func (i *inference) doHTTPExperimentInferenceRequest(ctx context.Context, experimentName, body string) error {
	return i.doHTTPInferenceRequest(ctx, experimentName, fmt.Sprintf("%s.experiment", experimentName), body)
}

func (i *inference) doHTTPPipelineInferenceRequest(ctx context.Context, pipelineName, body string) error {
	return i.doHTTPInferenceRequest(ctx, pipelineName, fmt.Sprintf("%s.pipeline", pipelineName), body)
}

func (i *inference) doHTTPModelInferenceRequest(ctx context.Context, modelName, body string) error {
	return i.doHTTPInferenceRequest(ctx, modelName, modelName, body)
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

	return i.doHTTPModelInferenceRequest(ctx, m.modelName, testModel.ValidHTTPInferenceRequest)
}

func httpScheme(useSSL bool) string {
	if useSSL {
		return "https"
	}
	return "http"
}

func (i *inference) sendGRPCModelInferenceRequest(ctx context.Context, model string, payload *godog.DocString) error {
	return i.doGRPCInferenceRequest(ctx, model, payload.Content)
}

func (i *inference) sendGRPCModelInferenceRequestFromModel(ctx context.Context, m *Model) error {
	testModel, ok := testModels[m.modelType]
	if !ok {
		return fmt.Errorf("could not find test model %s", m.model.Name)
	}
	return i.doGRPCInferenceRequest(ctx, m.modelName, testModel.ValidGRPCInferenceRequest)
}

func (i *inference) doGRPCInferenceRequest(
	ctx context.Context,
	resourceName string,
	payload string,
) error {
	var req v2_dataplane.ModelInferRequest
	if err := protojson.Unmarshal([]byte(payload), &req); err != nil {
		return fmt.Errorf("could not unmarshal gRPC json payload: %w", err)
	}
	req.ModelName = resourceName

	md := metadata.Pairs("seldon-model", resourceName)
	ctx = metadata.NewOutgoingContext(ctx, md)

	i.log.Debugf("sending gRPC model inference %+v", &req)

	resp, err := i.grpc.ModelInfer(ctx, &req)
	i.log.Debugf("grpc model infer response: %+v", resp)
	i.log.Debugf("grpc model infer error: %+v", err)

	i.lastGRPCResponse.response = resp
	i.lastGRPCResponse.err = err
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

func (i *inference) gRPCRespContainsNoError() error {
	if i.lastGRPCResponse.err != nil {
		return fmt.Errorf("grpc response contains error: %w", i.lastGRPCResponse.err)
	}
	if i.lastGRPCResponse.response == nil {
		return errors.New("grpc contains no response")
	}
	return nil
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
		if i.lastGRPCResponse.err != nil {
			return fmt.Errorf("no gRPC response, error found: %s", i.lastGRPCResponse.err.Error())
		}
		return errors.New("no gRPC response found")
	}

	gotJson, err := json.Marshal(i.lastGRPCResponse.response)
	if err != nil {
		return fmt.Errorf("could not marshal gRPC json: %w", err)
	}

	i.log.Debugf("checking gRPC response: %s contains %s", string(gotJson), expectJSON.Content)
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
	return i.doHttpRespCheckBodyContainsJSON(expectJSON.Content)
}

func (i *inference) doHttpRespCheckBodyContainsJSON(expectJSON string) error {
	if i.lastHTTPResponse == nil {
		return errors.New("no http response found")
	}

	body, err := io.ReadAll(i.lastHTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	i.log.Debugf("checking HTTP response: %s contains %s", string(body), expectJSON)
	ok, err := jsonContainsObjectSubset(string(body), expectJSON)
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
