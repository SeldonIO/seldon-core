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
