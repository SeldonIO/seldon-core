package client

import (
	"context"
	"fmt"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"golang.org/x/xerrors"
	"io"
)

const (
	SeldonPredictPath         = "/predict"
	SeldonTransformInputPath  = "/transform-input"
	SeldonTransformOutputPath = "/transform-output"
	SeldonCombinePath         = "/aggregate"
	SeldonRoutePath           = "/route"
	SeldonFeedbackPath        = "/send-feedback"
	SeldonStatusPath          = "/health/status"
	SeldonMetadataPath        = "/metadata"
)

type SeldonApiClient interface {
	Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error)
	TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error)
	Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (int, error)
	Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error)
	TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error)
	Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error)
	Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error)
	Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error)
	// Return model's metadata as payload.SeldonPaylaod (to expose as received on corresponding executor endpoint)
	Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error)
	// Return model's metadata decoded to payload.ModelMetadata (to build GraphMetadata)
	ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error)
	Unmarshall(msg []byte, contentType string) (payload.SeldonPayload, error)
	Marshall(out io.Writer, msg payload.SeldonPayload) error
	CreateErrorPayload(err error) payload.SeldonPayload
	IsGrpc() bool
}

type SeldonApiError struct {
	Message string
	Code    int
	frame   xerrors.Frame
}

func (se SeldonApiError) FormatError(p xerrors.Printer) error {
	p.Printf("%d %s", se.Code, se.Message)
	se.frame.Format(p)
	return nil
}

func (se SeldonApiError) Format(f fmt.State, c rune) {
	xerrors.FormatError(se, f, c)
}

func (se SeldonApiError) Error() string {
	return fmt.Sprint(se)
}
