package client

import (
	"fmt"
	"golang.org/x/xerrors"
	"io"
)

type SeldonApiClient interface {
	Predict(host string, port int32, msg SeldonPayload) (SeldonPayload, error)
	TransformInput(host string, port int32, msg SeldonPayload) (SeldonPayload, error)
	Route(host string, port int32, msg SeldonPayload) (int, error)
	Combine(host string, port int32, msgs []SeldonPayload) (SeldonPayload, error)
	TransformOutput(host string, port int32, msg SeldonPayload) (SeldonPayload, error)
	Unmarshall(msg []byte) (SeldonPayload, error)
	Marshall(out io.Writer, msg SeldonPayload) error
	CreateErrorPayload(err error) (SeldonPayload, error)
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
