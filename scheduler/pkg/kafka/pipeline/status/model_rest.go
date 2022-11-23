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

package status

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/pkg/util"
	"github.com/sirupsen/logrus"
)

type ModelRestCaller struct {
	envoyHost  string
	envoyPort  int
	logger     logrus.FieldLogger
	tlsOptions *util.TLSOptions
	httpClient *http.Client
}

func NewModelRestStatusCaller(logger logrus.FieldLogger, envoyHost string, envoyPort int) (*ModelRestCaller, error) {
	tlsOptions, err := util.CreateTLSClientOptions()
	if err != nil {
		return nil, err
	}
	httpClient := util.GetHttpClientFromTLSOptions(tlsOptions)
	if err != nil {
		return nil, err
	}
	return &ModelRestCaller{
		envoyHost:  envoyHost,
		envoyPort:  envoyPort,
		logger:     logger.WithField("source", "ModelRestCaller"),
		tlsOptions: tlsOptions,
		httpClient: httpClient,
	}, nil
}

func (mr *ModelRestCaller) getReadyUrl(modelName string) *url.URL {
	scheme := "http"
	if mr.tlsOptions.TLS {
		scheme = "https"
	}
	return &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(mr.envoyHost, strconv.Itoa(mr.envoyPort)),
		Path:   fmt.Sprintf("/v2/models/%s/ready", modelName),
	}
}

func (mr *ModelRestCaller) CheckModelReady(ctx context.Context, modelName string, requestId string) (bool, error) {
	logger := mr.logger.WithField("func", "CheckModelReady")
	url := mr.getReadyUrl(modelName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set(resources.SeldonModelHeader, modelName)
	req.Header.Set(util.RequestIdHeader, requestId)
	response, err := mr.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	logger.Infof("response from model %s %d", modelName, response.StatusCode)
	switch response.StatusCode {
	case http.StatusOK:
		return true, nil
	default:
		return false, nil
	}
}
