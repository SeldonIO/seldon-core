package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// we define all communication error into one bucket
	// TODO: separate out the different comm issues (e.g. DNS vs Connection refused etc.)
	HttpCommunicationErrCode = -100
	// i.e invalid method etc.
	HttpRequestErrCode = -200
)

type V2Client struct {
	host       string
	port       int
	httpClient *http.Client
	logger     log.FieldLogger
}

// Error wrapper with client and server errors + error code
// errCode should have the standard http error codes (for server)
// and client communication error codes (defined above)
type V2Err struct {
	err error
	// one bucket for http status code and client codes (for error)
	errCode int
}

type V2ServerError struct {
	Error string `json:"error"`
}

var ErrV2BadRequest = errors.New("V2 Bad Request")

func NewV2Client(host string, port int, logger log.FieldLogger) *V2Client {
	logger.Infof("V2 Inference Server %s:%d", host, port)

	netTransport := &http.Transport{
		MaxIdleConns:        maxIdleConnsHTTP,
		MaxIdleConnsPerHost: maxIdleConnsPerHostHTTP,
		DisableKeepAlives:   disableKeepAlivesHTTP,
		MaxConnsPerHost:     maxConnsPerHostHTTP,
	}
	netClient := &http.Client{
		Timeout:   time.Second * defaultTimeoutSeconds,
		Transport: netTransport,
	}

	return &V2Client{
		host:       host,
		port:       port,
		httpClient: netClient,
		logger:     logger.WithField("Source", "V2InferenceServerClient"),
	}
}

func (v *V2Client) getUrl(path string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(v.host, strconv.Itoa(v.port)),
		Path:   path,
	}
}

func (v *V2Client) call(path string) *V2Err {
	v2Url := v.getUrl(path)
	req, err := http.NewRequest("POST", v2Url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return &V2Err{
			err:     err,
			errCode: HttpRequestErrCode,
		}
	}
	response, err := v.httpClient.Do(req)
	if err != nil {
		return &V2Err{
			err:     err,
			errCode: HttpCommunicationErrCode,
		}
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &V2Err{
			err:     err,
			errCode: response.StatusCode,
		}
	}
	err = response.Body.Close()
	if err != nil {
		return &V2Err{
			err:     err,
			errCode: response.StatusCode,
		}
	}
	v.logger.Infof("v2 server response: %s", b)
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusBadRequest {
			v2Error := V2ServerError{}
			err := json.Unmarshal(b, &v2Error)
			if err != nil {
				return &V2Err{
					err:     err,
					errCode: response.StatusCode,
				}
			}
			return &V2Err{
				err:     fmt.Errorf("%s. %w", v2Error.Error, ErrV2BadRequest),
				errCode: response.StatusCode,
			}
		} else {
			return &V2Err{
				err:     fmt.Errorf("V2 server error: %s", b),
				errCode: response.StatusCode,
			}
		}
	}
	return nil
}

func (v *V2Client) LoadModel(name string) *V2Err {
	path := fmt.Sprintf("v2/repository/models/%s/load", name)
	v.logger.Infof("Load request: %s", path)
	return v.call(path)
}

func (v *V2Client) UnloadModel(name string) *V2Err {
	path := fmt.Sprintf("v2/repository/models/%s/unload", name)
	v.logger.Infof("Unload request: %s", path)
	return v.call(path)
}

func (v *V2Client) Ready() error {
	_, err := http.Get(v.getUrl("v2/health/ready").String())
	return err
}
