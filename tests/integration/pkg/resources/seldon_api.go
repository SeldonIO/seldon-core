/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package resources

import (
	"os"
	"time"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
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
	_, err = s.schedulerClient.Status(dat, true, 10*time.Second)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SeldonAPI) Infer(filename string, request string) ([]byte, error) {
	return s.inferClient.Infer(filename, request)
}
