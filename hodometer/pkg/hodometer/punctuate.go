/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package hodometer

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Punctuator struct {
	logger   logrus.FieldLogger
	interval time.Duration
}

func NewPunctuator(
	l logrus.FieldLogger,
	i time.Duration,
) *Punctuator {
	logger := l.WithField("source", "Punctuator")
	return &Punctuator{
		logger:   logger,
		interval: i,
	}
}

func (p *Punctuator) Run(description string, action func()) {
	logger := p.logger.WithField("func", "Run")

	runTime := time.Now().UTC()
	for {
		logger.Infof("%s at %v", description, runTime)
		action()
		logger.Infof("next run after %v", p.interval)
		runTime = <-time.After(p.interval)
	}
}
