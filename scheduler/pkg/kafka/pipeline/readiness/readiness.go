package readiness

import "fmt"

type Service interface {
	IsReady() error
	ID() string
}

type Manager interface {
	AddService(svc Service)
	IsReady() error
}

type manager struct {
	svcs []Service
}

func NewReadiness() Manager {
	return &manager{}
}

func (r *manager) AddService(svc Service) {
	r.svcs = append(r.svcs, svc)
}

func (r *manager) IsReady() error {
	for _, svc := range r.svcs {
		if err := svc.IsReady(); err != nil {
			return fmt.Errorf("service %s is not ready: %w", svc.ID(), err)
		}
	}
	return nil
}
