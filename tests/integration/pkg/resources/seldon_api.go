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

package resources

import (
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"os"
	"time"
)

type SeldonAPI struct {
	schedulerClient *cli.SchedulerClient
	inferClient     *SeldonInferAPI
}

func NewSeldonAPI() (*SeldonAPI, error) {
	// Setting hardwired ports for now
	sc, err := cli.NewSchedulerClient("0.0.0.0:9004", false, "", false)
	if err != nil {
		return nil, err
	}
	// Setting hardwired ports for now
	ic, err := NewSeldonInferAPI("0.0.0.0:9000")
	if err != nil {
		return nil, err
	}
	return &SeldonAPI{
		schedulerClient: sc,
		inferClient:     ic,
	}, nil
}

func (s *SeldonAPI) Load(filename string) error {
	dat, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = s.schedulerClient.Load(dat)
	if err != nil {
		return err
	}
	return nil
}

func (s *SeldonAPI) Unload(filename string) error {
	dat, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = s.schedulerClient.Unload(dat)
	return err
}

func (s *SeldonAPI) IsLoaded(filename string) (bool, error) {
	dat, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	_, err = s.schedulerClient.Status(dat, true, time.Duration(10*int64(time.Second)))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SeldonAPI) Infer(filename string, request string) ([]byte, error) {
	return s.inferClient.Infer(filename, request)
}
