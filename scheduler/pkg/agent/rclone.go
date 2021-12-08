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
	"strings"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	ContentTypeJSON        = "application/json"
	ContentType            = "Content-Type"
	RcloneNoopPath         = "/rc/noop"
	RcloneSyncCopyPath     = "/sync/copy"
	RcloneConfigCreatePath = "/config/create"
	RcloneConfigUpdatePath = "/config/update"
	RcloneListRemotesPath  = "/config/listremotes"
	RcloneConfigDeletePath = "/config/delete"
	RcloneConfigGetPath    = "/config/get"
)

type RCloneClient struct {
	host       string
	port       int
	localPath  string
	httpClient *http.Client
	logger     log.FieldLogger
	validate   *validator.Validate
}

type Noop struct {
	Foo string `json:"foo,omitempty" protobuf:"bytes,1,name=foo"`
}

type RcloneCopy struct {
	SrcFs              string `json:"srcFs"`
	DstFs              string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
}

type RcloneConfigKey struct {
	Name string `json:"name" yaml:"name"`
}

type RcloneConfigCreate struct {
	Name       string            `json:"name" yaml:"name" validate:"required"`
	Type       string            `json:"type" yaml:"type" validate:"required"`
	Parameters map[string]string `json:"parameters" yaml:"parameters" validate:"required"`
	Opt        map[string]string `json:"opt" yaml:"opt"`
}

type RcloneConfigUpdate struct {
	Name       string            `json:"name" yaml:"name"`
	Parameters map[string]string `json:"parameters" yaml:"parameters"`
	Opt        map[string]string `json:"opt" yaml:"opt"`
}

type RcloneListRemotes struct {
	Remotes []string `json:"remotes"`
}

type RcloneDeleteRemote struct {
	Name string `json:"name"`
}

func createConfigUpdateFromCreate(create *RcloneConfigCreate) *RcloneConfigUpdate {
	update := RcloneConfigUpdate{
		Name:       create.Name,
		Parameters: create.Parameters,
		Opt:        create.Opt,
	}
	return &update
}

func NewRCloneClient(host string, port int, localPath string, logger log.FieldLogger) *RCloneClient {
	logger.Infof("Rclone server %s:%d with model-repository:%s", host, port, localPath)
	return &RCloneClient{
		host:       host,
		port:       port,
		localPath:  localPath,
		httpClient: http.DefaultClient,
		logger:     logger.WithField("Source", "RCloneClient"),
		validate:   validator.New(),
	}
}

func (r *RCloneClient) call(op []byte, path string) ([]byte, error) {
	rcloneUrl := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(r.host, strconv.Itoa(r.port)),
		Path:   path,
	}
	r.logger.Infof("Calling Rclone server: %s with %s", path, string(op))
	req, err := http.NewRequest("POST", rcloneUrl.String(), bytes.NewBuffer(op))
	if err != nil {
		return nil, err
	}
	req.Header.Add(ContentType, ContentTypeJSON)
	response, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = response.Body.Close()
	if err != nil {
		return nil, err
	}
	r.logger.Printf("rclone response: %s", b)
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed rclone request to host:%s port:%d path:%s", r.host, r.port, path)
	}
	return b, nil
}

func (r *RCloneClient) Ready() error {
	noop := Noop{Foo: "bar"}
	b, err := json.Marshal(noop)
	if err != nil {
		return err
	}
	_, err = r.call(b, RcloneNoopPath)
	return err
}

// This method assumes a simple remote URI with no config. e.g. s3://mybucket
func getRemoteName(uri string) (string, error) {
	idx := strings.Index(uri, ":")
	if idx == -1 {
		return "", fmt.Errorf("Failed to find : in %s for rclone name match", uri)
	}
	if idx == 0 {
		return "", fmt.Errorf("Can't get remote from URI with configuration included inline")
	}
	name := uri[0:idx]
	return name, nil
}

func (r *RCloneClient) parseRcloneConfig(config []byte) (*RcloneConfigCreate, error) {
	configCreate := RcloneConfigCreate{}
	err := json.Unmarshal(config, &configCreate)
	if err != nil {
		err2 := yaml.Unmarshal(config, &configCreate)
		if err2 != nil {
			return nil, fmt.Errorf("Failed to unmarshall config as json or yaml. JSON error %s. YAML error %s", err.Error(), err2.Error())
		}
	}
	err = r.validate.Struct(configCreate)
	if err != nil {
		return nil, err
	}
	return &configCreate, nil
}

// Creating a connection string with https://rclone.org/docs/#connection-strings
func (r *RCloneClient) createUriWithConfig(uri string, config []byte) (string, error) {
	remote, err := getRemoteName(uri)
	if err != nil {
		return "", err
	}
	parsed, err := r.parseRcloneConfig(config)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString(":")
	sb.WriteString(remote)
	for k, v := range parsed.Parameters {
		sb.WriteString(",")
		sb.WriteString(k)
		sb.WriteString("=")
		if strings.ContainsAny(v, ":,") {
			sb.WriteString(`"`)
			v = strings.Replace(v, `"`, `""`, -1)
		}
		sb.WriteString(v)
		if strings.ContainsAny(v, ":,") {
			sb.WriteString(`"`)
		}
	}
	return strings.Replace(uri, remote, sb.String(), 1), nil
}

func (r *RCloneClient) Config(config []byte) (string, error) {
	configCreate, err := r.parseRcloneConfig(config)
	if err != nil {
		return "", err
	}
	exists, err := r.configExists(configCreate.Name)
	if err != nil {
		return "", err
	}
	if exists {
		return configCreate.Name, r.configUpdate(configCreate)
	} else {
		return configCreate.Name, r.configCreate(configCreate)
	}
}

// Call Rclone /sync/copy
func (r *RCloneClient) Copy(modelName string, src string, config []byte) error {
	var srcUpdated string
	var err error
	if len(config) > 0 {
		srcUpdated, err = r.createUriWithConfig(src, config)
		if err != nil {
			return err
		}
	} else {
		srcUpdated = src
	}

	dst := fmt.Sprintf("%s/%s", r.localPath, modelName)
	copy := RcloneCopy{
		SrcFs:              srcUpdated,
		DstFs:              dst,
		CreateEmptySrcDirs: true,
	}
	r.logger.Infof("Copy from %s (original %s) to %s", srcUpdated, src, dst)
	b, err := json.Marshal(copy)
	if err != nil {
		return err
	}
	_, err = r.call(b, RcloneSyncCopyPath)
	return err
}

// Call Rclone /config/get
func (r *RCloneClient) configExists(rcloneRemoteKey string) (bool, error) {
	key := RcloneConfigKey{Name: rcloneRemoteKey}
	b, err := json.Marshal(key)
	if err != nil {
		return false, err
	}
	res, err := r.call(b, RcloneConfigGetPath)
	if err != nil {
		return false, err
	}
	var anyJson map[string]interface{}
	err = json.Unmarshal(res, &anyJson)
	if err != nil {
		return false, err
	}
	if len(anyJson) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// Call Rclone /config/create
func (r *RCloneClient) configCreate(configCreate *RcloneConfigCreate) error {
	b, err := json.Marshal(configCreate)
	if err != nil {
		return err
	}
	_, err = r.call(b, RcloneConfigCreatePath)
	return err
}

// Call Rclone /config/update
func (r *RCloneClient) configUpdate(configCreate *RcloneConfigCreate) error {
	configUpdate := createConfigUpdateFromCreate(configCreate)
	b, err := json.Marshal(configUpdate)
	if err != nil {
		return err
	}
	_, err = r.call(b, RcloneConfigUpdatePath)
	return err
}

// Call Rclone /config/listremotes
func (r *RCloneClient) ListRemotes() ([]string, error) {
	res, err := r.call([]byte("{}"), RcloneListRemotesPath)
	if err != nil {
		return nil, err
	}
	remotes := RcloneListRemotes{}
	err = json.Unmarshal(res, &remotes)
	if err != nil {
		return nil, err
	}
	return remotes.Remotes, nil
}

func (r *RCloneClient) DeleteRemote(name string) error {
	delRemote := RcloneDeleteRemote{Name: name}
	b, err := json.Marshal(&delRemote)
	if err != nil {
		return err
	}
	_, err = r.call(b, RcloneConfigDeletePath)
	return err
}
