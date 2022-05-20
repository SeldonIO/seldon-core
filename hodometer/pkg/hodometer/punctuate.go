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
