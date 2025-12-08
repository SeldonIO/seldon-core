package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
)

func (i *inference) sendHTTPModelInferenceRequest(timeout, model string, payload *godog.DocString) error {
	ctx, cancel, err := createTimeoutCtx(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout %s: %w", timeout, err)
	}
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/v2/models/%s/infer", i.host, model), strings.NewReader(payload.Content))
	if err != nil {
		return fmt.Errorf("could not create http request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", "seldon-mesh.inference.seldon")
	req.Header.Add("Seldon-model", model)

	resp, err := i.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not send http request: %w", err)
	}
	i.lastHTTPResponse = resp
	return nil
}

func (i *inference) sendGRPCModelInferenceRequest(timeout, model string, payload *godog.DocString) error {
	ctx, cancel, err := createTimeoutCtx(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout %s: %w", timeout, err)
	}
	defer cancel()

	var msg *v2_dataplane.ModelInferRequest
	if err := json.Unmarshal([]byte(payload.Content), &msg); err != nil {
		return fmt.Errorf("could not unmarshal gRPC json payload: %w", err)
	}
	msg.ModelName = model

	resp, err := i.grpc.ModelInfer(ctx, msg)
	if err != nil {
		return fmt.Errorf("could not send grpc model inference: %w", err)
	}

	i.lastGRPCResponse = resp
	return nil
}

func createTimeoutCtx(timeout string) (context.Context, context.CancelFunc, error) {
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid timeout %s: %w", timeout, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	return ctx, cancel, nil
}
