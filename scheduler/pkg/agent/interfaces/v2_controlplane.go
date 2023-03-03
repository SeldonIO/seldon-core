/*
Copyright 2022 Seldon Technologies Ltd.

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

	"google.golang.org/grpc/codes"
)

// Error wrapper with client and server errors + error code
// errCode should have the standard http error codes (for server)
// and client communication error codes (defined above)
type V2Err struct {
	IsGrpc bool
	Err    error
	// one bucket for http/grpc status code and client codes (for error)
	ErrCode int
}

func (v2err *V2Err) Error() string {
	if v2err != nil {
		return v2err.Err.Error()
	}
	return ""
}

func (v *V2Err) IsNotFound() bool {
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

var ErrV2BadRequest = errors.New("V2 Bad Request")
var ErrServerNotReady = errors.New("Server not ready")

type V2Client interface {
	LoadModel(name string) *V2Err
	UnloadModel(name string) *V2Err
	Live() error
	GetModels() ([]ServerModelInfo, error)
}
