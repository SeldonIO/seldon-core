package resources

import (
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"os"
)

type SeldonAPI struct {
	schedulerClient *cli.SchedulerClient
	inferClient     *SeldonInferAPI
}

func NewSeldonAPI() (*SeldonAPI, error) {
	sc, err := cli.NewSchedulerClient("", false, "")
	if err != nil {
		return nil, err
	}
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

func (s *SeldonAPI) UnLoad(filename string) error {
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
	_, err = s.schedulerClient.Status(dat, true, 10)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SeldonAPI) Infer(filename string, request string) error {
	return s.inferClient.Infer(filename, request)
}
