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

package testing_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type V2RestClientForTest struct {
	HttpClient *http.Client
	host       string
	httpPort   int
	logger     log.FieldLogger
}

func NewV2RestClientForTest(host string, port int, logger log.FieldLogger) *V2RestClientForTest {
	logger.Infof("V2 Inference Server %s:%d", host, port)

	netTransport := &http.Transport{
		MaxIdleConns:        util.MaxIdleConnsHTTP,
		MaxIdleConnsPerHost: util.MaxIdleConnsPerHostHTTP,
		DisableKeepAlives:   util.DisableKeepAlivesHTTP,
		MaxConnsPerHost:     util.MaxConnsPerHostHTTP,
		IdleConnTimeout:     util.IdleConnTimeoutSeconds * time.Second,
	}
	netClient := &http.Client{
		Timeout:   time.Second * util.DefaultTimeoutSeconds,
		Transport: netTransport,
	}

	return &V2RestClientForTest{
		host:       host,
		httpPort:   port,
		HttpClient: netClient,
		logger:     logger.WithField("Source", "V2InferenceServerClientHttp"),
	}

}

func (v *V2RestClientForTest) getUrl(path string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(v.host, strconv.Itoa(v.httpPort)),
		Path:   path,
	}
}

func (v *V2RestClientForTest) call(path string) *interfaces.ControlPlaneErr {
	v2Url := v.getUrl(path)
	req, err := http.NewRequest("POST", v2Url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return &interfaces.ControlPlaneErr{
			IsGrpc:  false,
			Err:     err,
			ErrCode: interfaces.V2RequestErrCode,
		}
	}
	response, err := v.HttpClient.Do(req)
	if err != nil {
		return &interfaces.ControlPlaneErr{
			IsGrpc:  false,
			Err:     err,
			ErrCode: interfaces.V2CommunicationErrCode,
		}
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return &interfaces.ControlPlaneErr{
			IsGrpc:  false,
			Err:     err,
			ErrCode: response.StatusCode,
		}
	}
	err = response.Body.Close()
	if err != nil {
		return &interfaces.ControlPlaneErr{
			IsGrpc:  false,
			Err:     err,
			ErrCode: response.StatusCode,
		}
	}
	v.logger.Infof("v2 server response: %s", b)
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusBadRequest {
			v2Error := interfaces.V2ServerError{}
			err := json.Unmarshal(b, &v2Error)
			if err != nil {
				return &interfaces.ControlPlaneErr{
					IsGrpc:  false,
					Err:     err,
					ErrCode: response.StatusCode,
				}
			}
			return &interfaces.ControlPlaneErr{
				IsGrpc:  false,
				Err:     fmt.Errorf("%s. %w", v2Error.Error, interfaces.ErrControlPlaneBadRequest),
				ErrCode: response.StatusCode,
			}
		} else {
			return &interfaces.ControlPlaneErr{
				IsGrpc:  false,
				Err:     fmt.Errorf("V2 server error: %s", b),
				ErrCode: response.StatusCode,
			}
		}
	}
	return nil
}

func (v *V2RestClientForTest) LoadModel(name string) *interfaces.ControlPlaneErr {
	return v.loadModelHttp(name)
}

func (v *V2RestClientForTest) loadModelHttp(name string) *interfaces.ControlPlaneErr {
	path := fmt.Sprintf("v2/repository/models/%s/load", name)
	v.logger.Infof("Load request: %s", path)
	return v.call(path)
}

func (v *V2RestClientForTest) UnloadModel(name string) *interfaces.ControlPlaneErr {
	return v.unloadModelHttp(name)
}

func (v *V2RestClientForTest) unloadModelHttp(name string) *interfaces.ControlPlaneErr {
	path := fmt.Sprintf("v2/repository/models/%s/unload", name)
	v.logger.Infof("Unload request: %s", path)
	return v.call(path)
}

func (v *V2RestClientForTest) Live() error {
	var ready bool
	var err error

	ready, err = v.liveHttp()

	if err != nil {
		v.logger.WithError(err).Debugf("Server live check failed on error")
		return err
	}
	if ready {
		return nil
	} else {
		return interfaces.ErrServerNotReady
	}
}

func (v *V2RestClientForTest) liveHttp() (bool, error) {
	res, err := http.Get(v.getUrl("v2/health/live").String())
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, nil
	}
}

func (v *V2RestClientForTest) GetModels() ([]interfaces.ServerModelInfo, error) {
	v.logger.Warnf("Http GetModels not available returning empty list")
	return []interfaces.ServerModelInfo{}, nil
}
