/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
