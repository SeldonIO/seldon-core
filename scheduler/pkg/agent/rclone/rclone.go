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

package rclone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	ContentTypeJSON           = "application/json"
	ContentType               = "Content-Type"
	RcloneNoopPath            = "/rc/noop"
	RcloneSyncCopyPath        = "/sync/copy"
	RcloneOperationsPurgePath = "/operations/purge"
	RcloneConfigCreatePath    = "/config/create"
	RcloneConfigUpdatePath    = "/config/update"
	RcloneListRemotesPath     = "/config/listremotes"
	RcloneConfigDeletePath    = "/config/delete"
	RcloneConfigGetPath       = "/config/get"
)

type RCloneClient struct {
	host       string
	port       int
	localPath  string
	httpClient *http.Client
	logger     log.FieldLogger
	validate   *validator.Validate
	namespace  string
	configChan chan config.AgentConfiguration
}

type Noop struct {
	Foo string `json:"foo,omitempty" protobuf:"bytes,1,name=foo"`
}

type RcloneCopy struct {
	SrcFs              string `json:"srcFs"`
	DstFs              string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
}

type RclonePurge struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
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

func NewRCloneClient(
	host string,
	port int,
	localPath string,
	logger log.FieldLogger,
	namespace string,
) *RCloneClient {
	logger.Infof("Rclone server %s:%d with model-cache:%s", host, port, localPath)
	return &RCloneClient{
		host:       host,
		port:       port,
		localPath:  localPath,
		httpClient: http.DefaultClient,
		logger:     logger.WithField("Source", "RCloneClient"),
		validate:   validator.New(),
		namespace:  namespace,
		configChan: make(chan config.AgentConfiguration),
	}
}

func (r *RCloneClient) StartConfigListener(configHandler *config.AgentConfigHandler) error {
	logger := r.logger.WithField("func", "StartConfigListener")
	// Start config listener
	go r.listenForConfigUpdates()
	// Add ourself as listener on channel and handle initial config
	logger.Info("Loading initial rclone configuration")
	err := r.loadRcloneConfiguration(configHandler.AddListener(r.configChan))
	if err != nil {
		logger.WithError(err).Errorf("Failed to load rclone defaults")
		return err
	}
	return nil
}

func (r *RCloneClient) listenForConfigUpdates() {
	logger := r.logger.WithField("func", "listenForConfigUpdates")
	for config := range r.configChan {
		logger.Info("Received config update")
		config := config
		go func() {
			err := r.loadRcloneConfiguration(&config)
			if err != nil {
				logger.WithError(err).Error("Failed to load rclone defaults")
			}
		}()
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
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = response.Body.Close()
	if err != nil {
		return nil, err
	}

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
func (r *RCloneClient) createUriWithConfig(uri string, rawConfig []byte) (string, error) {
	remoteName, err := getRemoteName(uri)
	if err != nil {
		return "", err
	}

	config, err := r.parseRcloneConfig(rawConfig)
	if err != nil {
		return "", err
	}

	if config.Name != remoteName {
		return "", fmt.Errorf(
			"name from URI (%s) does not match secret (%s); are you using the right storage config?",
			remoteName,
			config.Name,
		)
	}

	var sb strings.Builder
	sb.WriteString(":")
	sb.WriteString(config.Type)
	for k, v := range config.Parameters {
		sb.WriteString(",")
		sb.WriteString(k)
		sb.WriteString("=")

		if strings.ContainsAny(v, ":,") {
			sb.WriteString(`"`)
			sb.WriteString(
				strings.Replace(v, `"`, `""`, -1),
			)
			sb.WriteString(`"`)
		} else {
			sb.WriteString(v)
		}
	}

	return strings.Replace(uri, remoteName, sb.String(), 1), nil
}

func (r *RCloneClient) Config(config []byte) (string, error) {
	logger := r.logger.WithField("func", "Config")

	configCreate, err := r.parseRcloneConfig(config)
	if err != nil {
		return "", err
	}

	logger.WithField("remote name", configCreate.Name).Info("loaded config")

	exists, err := r.configExists(configCreate.Name)
	if err != nil {
		return "", err
	}

	if exists {
		logger.WithField("remote name", configCreate.Name).Info("updating existing Rclone remote")
		return configCreate.Name, r.configUpdate(configCreate)
	} else {
		logger.WithField("remote name", configCreate.Name).Info("creating new Rclone remote")
		return configCreate.Name, r.configCreate(configCreate)
	}
}

func CreateRcloneModelHash(modelName string, srcUri string) (uint32, error) {
	keyTohash := modelName + "-" + srcUri
	return util.Hash(keyTohash)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Call Rclone /sync/copy
func (r *RCloneClient) Copy(modelName string, srcUri string, config []byte) (string, error) {
	logger := r.logger.WithField("func", "Copy")

	var srcUpdated string
	var err error
	if len(config) > 0 {
		srcUpdated, err = r.createUriWithConfig(srcUri, config)
		if err != nil {
			return "", err
		}
	} else {
		srcUpdated = srcUri
	}

	// Create key from srcUri plus modelName.
	// If we just used srcUri it would mean models with same srcUri could share rclone download
	// However, its unclear whether we might get partial updates under load if srcUri changes and two
	// or more models are asking for the uri.
	// TODO  reinvestigate how we can use rclone sharing maybe via an rclone proxy caching layer?
	hash, err := CreateRcloneModelHash(modelName, srcUri)
	if err != nil {
		return "", err
	}

	dst := fmt.Sprintf("%s/%d", r.localPath, hash)
	copy := RcloneCopy{
		SrcFs:              srcUpdated,
		DstFs:              dst,
		CreateEmptySrcDirs: true,
	}

	logger.
		WithField("source", srcUri).
		WithField("destination", dst).
		Info("will copy model artifacts")

	b, err := json.Marshal(copy)
	if err != nil {
		return "", err
	}

	_, err = r.call(b, RcloneSyncCopyPath)
	if err != nil {
		return "", fmt.Errorf("Failed to sync/copy %s to %s %w", srcUri, dst, err)
	}

	// Even if we had success from rclone the src may be empty so need to check
	pathExists, err := pathExists(dst)
	if err != nil {
		return "", err
	}
	if !pathExists {
		return "", fmt.Errorf("Failed to download from %s any files", srcUri)
	}

	return dst, nil
}

func (r *RCloneClient) PurgeLocal(path string) error {
	p := RclonePurge{
		Fs:     path,
		Remote: "",
	}
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = r.call(b, RcloneOperationsPurgePath)
	if err != nil {
		return fmt.Errorf("Failed to run %s for %s %w", RcloneOperationsPurgePath, path, err)
	}
	return nil
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
