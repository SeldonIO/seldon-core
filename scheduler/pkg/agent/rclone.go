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
	ContentTypeJSON = "application/json"
	ContentType     = "Content-Type"
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

type RCloneCopy struct {
	SrcFs              string `json:"srcFs"`
	DstFs              string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
}

type RCloneConfigKey struct {
	Name string `json:"name" yaml:"name"`
}

type RCloneConfigCreate struct {
	Name       string            `json:"name" yaml:"name" validate:"required"`
	Type       string            `json:"type" yaml:"type" validate:"required"`
	Parameters map[string]string `json:"parameters" yaml:"parameters" validate:"required"`
	Opts       map[string]string `json:"opts" yaml:"opts"`
}

type RCloneConfigUpdate struct {
	Name       string            `json:"name" yaml:"name"`
	Parameters map[string]string `json:"parameters" yaml:"parameters"`
	Opts       map[string]string `json:"opts" yaml:"opts"`
}

func createConfigUpdateFromCreate(create *RCloneConfigCreate) *RCloneConfigUpdate {
	update := RCloneConfigUpdate{
		Name:       create.Name,
		Parameters: create.Parameters,
		Opts:       create.Opts,
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
	_, err = r.call(b, "/rc/noop")
	return err
}

func getRemoteName(uri string) (string, error) {
	idx := strings.Index(uri, ":")
	if idx == -1 {
		return "", fmt.Errorf("Failed to find : in %s for rclone name match", uri)
	}
	name := uri[0:idx]
	return name, nil
}

func createRCloneKey(modelName string, modelVersion string, remoteName string) string {
	return modelName + "-" + modelVersion + "-" + remoteName
}

func updatePath(srcPath string, currentRemoteKey string, rcloneUniqueName string) string {
	return strings.Replace(srcPath, currentRemoteKey, rcloneUniqueName, 1)
}

func (r *RCloneClient) Copy(modelName string, modelVersion string, src string) error {
	currentRemoteKey, err := getRemoteName(src)
	if err != nil {
		return err
	}
	rcloneRemoteKey := createRCloneKey(modelName, modelVersion, currentRemoteKey)
	srcUpdated := updatePath(src, currentRemoteKey, rcloneRemoteKey)
	dst := fmt.Sprintf("%s/%s", r.localPath, modelName)
	copy := RCloneCopy{
		SrcFs:              srcUpdated,
		DstFs:              dst,
		CreateEmptySrcDirs: true,
	}
	r.logger.Infof("Copy from %s (original %s) to %s", srcUpdated, src, dst)
	b, err := json.Marshal(copy)
	if err != nil {
		return err
	}
	_, err = r.call(b, "/sync/copy")
	return err
}

func (r *RCloneClient) Config(modelName string, modelVersion string, config []byte) error {
	logger := r.logger.WithField("func", "Config")
	configCreate, err := r.createConfigCreate(modelName, modelVersion, config)
	if err != nil {
		return err
	}
	exists, err := r.configExists(modelName, modelVersion, configCreate.Name)
	if err != nil {
		return err
	}
	if exists {
		logger.Infof("Config exists for %s:%s", modelName, modelVersion)
		return r.configUpdate(modelName, modelVersion, configCreate)
	} else {
		logger.Infof("Config does not exists for %s:%s", modelName, modelVersion)
		return r.configCreate(modelName, modelVersion, configCreate)
	}
}

func (r *RCloneClient) configExists(modelName string, modelVersion string, rcloneRemoteKey string) (bool, error) {
	key := RCloneConfigKey{Name: rcloneRemoteKey}
	b, err := json.Marshal(key)
	if err != nil {
		return false, err
	}
	res, err := r.call(b, "/config/get")
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

func (r *RCloneClient) createConfigCreate(modelName string, modelVersion string, config []byte) (*RCloneConfigCreate, error) {
	configCreate := RCloneConfigCreate{}
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
	configCreate.Name = createRCloneKey(modelName, modelVersion, configCreate.Name) //overwrite name with model name and version which is unique?
	return &configCreate, nil
}

func (r *RCloneClient) configCreate(modelName string, modelVersion string, configCreate *RCloneConfigCreate) error {
	logger := r.logger.WithField("func", "ConfigCreate")
	logger.Infof("model %s version %s", modelName, modelVersion)
	b, err := json.Marshal(configCreate)
	if err != nil {
		return err
	}
	_, err = r.call(b, "/config/create")
	return err
}

func (r *RCloneClient) configUpdate(modelName string, modelVersion string, configCreate *RCloneConfigCreate) error {
	logger := r.logger.WithField("func", "ConfigUpdate")
	logger.Infof("model %s version %s", modelName, modelVersion)
	configUpdate := createConfigUpdateFromCreate(configCreate)
	b, err := json.Marshal(configUpdate)
	if err != nil {
		return err
	}
	_, err = r.call(b, "/config/update")
	return err
}