package heartbeat

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

//go:generate go tool mockgen -source=./manager.go -destination=./mock/manager.go -package=mock Service
type Service interface {
	Name() string
	RequestHeartBeat() chan<- struct{}
	Heartbeat() <-chan error
}

type Manager struct {
	svcs []Service
	log  *logrus.Logger
}

func NewManager(log *logrus.Logger) *Manager {
	return &Manager{
		svcs: make([]Service, 0),
		log:  log,
	}
}

func (m *Manager) Register(svcs ...Service) {
	m.svcs = append(m.svcs, svcs...)
}

func (m *Manager) CheckHeartbeats(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)

	for _, svc := range m.svcs {
		errGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case svc.RequestHeartBeat() <- struct{}{}:
				select {
				case <-ctx.Done():
					return ctx.Err()
				case err, ok := <-svc.Heartbeat():
					m.log.WithField("service", svc.Name()).Debug("Heartbeat: got channel response")
					if !ok {
						return fmt.Errorf("heartbeat channel closed for %s", svc.Name())
					}
					if err != nil {
						return fmt.Errorf("heartbeat failed for %s: %w", svc.Name(), err)
					}
				}
			}

			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("failed waiting for heartbeats: %w", err)
	}
	m.log.Debug("Successfully checked all heartbeats")
	return nil
}
