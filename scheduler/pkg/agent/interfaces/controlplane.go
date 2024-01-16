/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package interfaces

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

// Error wrapper with client and server errors + error code
// errCode should have the standard http error codes (for server)
// and client communication error codes (defined above)
type ControlPlaneErr struct {
	IsGrpc bool
	Err    error
	// one bucket for http/grpc status code and client codes (for error)
	ErrCode int
}

func (err *ControlPlaneErr) Error() string {
	if err != nil {
		return err.Err.Error()
	}
	return ""
}

func (v *ControlPlaneErr) IsNotFound() bool {
	if v.IsGrpc {
		return v.ErrCode == int(codes.NotFound)
	} else {
		// assumes http
		return v.ErrCode == http.StatusNotFound
	}
}

type ServerModelState string

const (
	ServerModelState_UNKNOWN     ServerModelState = "UNKNOWN"
	ServerModelState_READY       ServerModelState = "READY"
	ServerModelState_UNAVAILABLE ServerModelState = "UNAVAILABLE"
	ServerModelState_LOADING     ServerModelState = "LOADING"
	ServerModelState_UNLOADING   ServerModelState = "UNLOADING"
)

const (
	// we define all communication error into one bucket
	// TODO: separate out the different comm issues (e.g. DNS vs Connection refused etc.)
	V2CommunicationErrCode = -100
	// i.e invalid method etc.
	V2RequestErrCode = -200
)

type ServerModelInfo struct {
	Name  string
	State ServerModelState
}

type V2ServerError struct {
	Error string `json:"error"`
}

type ModelServerConfig struct {
	Host   string
	Port   int
	Logger log.FieldLogger
}

var ErrControlPlaneBadRequest = errors.New("ControlPlane Bad Request")
var ErrServerNotReady = errors.New("Server not ready")

type ModelServerControlPlaneClient interface {
	LoadModel(name string) *ControlPlaneErr
	UnloadModel(name string) *ControlPlaneErr
	Live() error
	GetModels() ([]ServerModelInfo, error)
}
