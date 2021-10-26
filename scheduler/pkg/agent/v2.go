package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type V2Client struct {
	host string
	port int
	httpClient     *http.Client
	logger log.FieldLogger
}

type V2Error struct {
	Error string `json:"error"`
}

var V2BadRequestErr = errors.New("V2 Bad Request")

func NewV2Client(host string, port int, logger log.FieldLogger) *V2Client {
	logger.Infof("V2 Inference Server %s:%d",host, port)
	return &V2Client{
		host: host,
		port: port,
		httpClient: http.DefaultClient,
		logger: logger.WithField("Source","V2Client"),
	}
}

func (v *V2Client) getUrl(path string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(v.host, strconv.Itoa(v.port)),
		Path:   path,
	}
}

func (v *V2Client) call(path string) error {
	v2Url := v.getUrl(path)
	req, err := http.NewRequest("POST", v2Url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	response, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = response.Body.Close()
	if err != nil {
		return err
	}
	v.logger.Infof("v2 server response: %s",b)
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusBadRequest {
			v2Error := V2Error{}
			err := json.Unmarshal(b, &v2Error)
			if err != nil {
				return err
			}
			return fmt.Errorf("%s. %w",v2Error.Error, V2BadRequestErr)
		} else {
			return err
		}
	}
	return nil
}

func (v *V2Client) LoadModel(name string) error {
	path := fmt.Sprintf("v2/repository/models/%s/load",name)
	v.logger.Infof("Load request: %s",path)
	return v.call(path)
}

func (v *V2Client) UnloadModel(name string) error {
	path := fmt.Sprintf("v2/repository/models/%s/unload",name)
	v.logger.Infof("Unload request: %s",path)
	return v.call(path)
}

func (v *V2Client) Ready() error {
	_,err := http.Get(v.getUrl("v2/health/ready").String())
	return err
}