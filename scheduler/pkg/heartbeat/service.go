package heartbeat

import (
	"context"
)

type Check func() error

type Svc struct {
	name      string
	heartbeat Check
	result    chan error
	request   chan struct{}
}

var _ Service = &Svc{}

func NewService(name string, heartbeat Check) *Svc {
	return &Svc{
		name:      name,
		heartbeat: heartbeat,
		result:    make(chan error),
		request:   make(chan struct{}),
	}
}

func (s *Svc) RequestHeartBeat() chan<- struct{} {
	return s.request
}

func (s *Svc) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.request:
			s.result <- s.heartbeat()
		}
	}
}

func (s *Svc) Name() string {
	return s.name
}

func (s *Svc) Heartbeat() <-chan error {
	return s.result
}
