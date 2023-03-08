/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
