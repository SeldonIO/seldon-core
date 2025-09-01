package health_probe

import (
	"fmt"
	"slices"
)

type ProbeType int

const (
	ProbeStartUp ProbeType = 1 << iota
	ProbeReadiness
	ProbeLiveness
)

type Manager interface {
	AddCheck(cb ProbeCallback, probes ...ProbeType)
	CheckReadiness() error
	CheckLiveness() error
	CheckStartup() error
}

type ProbeCallback func() error

func (p ProbeType) Valid() bool {
	return slices.Contains([]ProbeType{ProbeStartUp, ProbeReadiness, ProbeLiveness}, p)
}

type manager struct {
	svcs []service
}

type service struct {
	probes ProbeType
	cb     ProbeCallback
}

func NewManager() Manager {
	return &manager{
		svcs: make([]service, 0),
	}
}

func (m *manager) AddCheck(cb ProbeCallback, probes ...ProbeType) {
	if cb == nil {
		panic("nil callback")
	}

	var probesSum ProbeType
	for _, probe := range probes {
		if !probe.Valid() {
			panic(fmt.Sprintf("Invalid probe type %v", probe))
		}
		probesSum |= probe
	}

	m.svcs = append(m.svcs, service{
		probes: probesSum,
		cb:     cb,
	})
}

func (m *manager) CheckReadiness() error {
	return m.runCheck(ProbeReadiness)
}

func (m *manager) CheckStartup() error {
	return m.runCheck(ProbeStartUp)
}

func (m *manager) CheckLiveness() error {
	return m.runCheck(ProbeLiveness)
}

func (m *manager) runCheck(probe ProbeType) error {
	for _, svc := range m.svcs {
		if svc.probes&probe == 0 {
			continue
		}
		if err := svc.cb(); err != nil {
			return fmt.Errorf("failed probe: %w", err)
		}
	}
	return nil
}
