package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	ContentTypeJSON = "application/json"
	ContentType     = "Content-Type"
)

type RCloneClient struct {
	host       string
	port       int
	localPath  string
	httpClient *http.Client
	logger     log.FieldLogger
}

type Noop struct {
	Foo string `json:"foo,omitempty" protobuf:"bytes,1,name=foo"`
}

type RcloneCopy struct {
	SrcFs              string `json:"srcFs"`
	DstFs              string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
}

func NewRCloneClient(host string, port int, localPath string, logger log.FieldLogger) *RCloneClient {
	logger.Infof("Rclone server %s:%d with model-repository:%s", host, port, localPath)
	return &RCloneClient{
		host:       host,
		port:       port,
		localPath:  localPath,
		httpClient: http.DefaultClient,
		logger:     logger.WithField("Source", "RCloneClient"),
	}
}

func (r *RCloneClient) call(op []byte, path string) error {
	rcloneUrl := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(r.host, strconv.Itoa(r.port)),
		Path:   path,
	}

	req, err := http.NewRequest("POST", rcloneUrl.String(), bytes.NewBuffer(op))
	if err != nil {
		return err
	}
	req.Header.Add(ContentType, ContentTypeJSON)
	response, err := r.httpClient.Do(req)
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
	r.logger.Printf("rclone response: %s", b)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed rclone request to host:%s port:%d path:%s", r.host, r.port, path)
	}
	return nil
}

func (r *RCloneClient) Ready() error {
	noop := Noop{Foo: "bar"}
	b, err := json.Marshal(noop)
	if err != nil {
		return err
	}
	return r.call(b, "/rc/noop")
}

func (r *RCloneClient) Copy(modelName string, src string) error {
	dst := fmt.Sprintf("%s/%s", r.localPath, modelName)
	copy := RcloneCopy{
		SrcFs:              src,
		DstFs:              dst,
		CreateEmptySrcDirs: true,
	}
	r.logger.Infof("Copy from %s to %s", src, dst)
	b, err := json.Marshal(copy)
	if err != nil {
		return err
	}
	return r.call(b, "/sync/copy")
}
