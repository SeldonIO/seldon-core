/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package status

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
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
	req.Header.Set(util.SeldonModelHeader, modelName)
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
