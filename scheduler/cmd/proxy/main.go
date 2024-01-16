/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	proxy "github.com/seldonio/seldon-core/scheduler/v2/pkg/proxy"
)

func main() {
	config := proxy.GetConfig()

	l := logrus.New()
	l.SetLevel(config.LogLevel)

	l.Infof("configuration: %+v", config)

	done := make(chan struct{})

	go func() {
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)

		s := <-term

		l.Infof("received %v signal", s)
		close(done)
	}()

	agentServer := proxy.NewAgentServer(
		l.WithFields(
			logrus.Fields{
				"component": "agent",
			},
		),
	)

	modelEvents := make(chan proxy.ModelEvent, config.ProxyEventBufferSize)
	proxyServer := proxy.NewProxyServer(
		l.WithFields(
			logrus.Fields{
				"component": "proxy",
			},
		),
		modelEvents,
	)
	scheduler := proxy.NewScheduler(
		l.WithFields(
			logrus.Fields{
				"component": "scheduler",
			},
		),
		modelEvents,
		agentServer,
	)

	proxyChainer := proxy.New(l)
	go func() {
		err := proxyChainer.Start(config.ChainerListenPort)
		if err != nil {
			l.WithError(err).Errorf("chainer server failed")
		}

		close(done)
	}()

	go func() {
		err := agentServer.Start(config.AgentListenPort)
		if err != nil {
			l.WithError(err).Errorf("agent server failed")
		}

		close(done)
	}()

	go scheduler.Start()

	go func() {
		err := proxyServer.Start(config.ProxyListenPort)
		l.WithError(err).Errorf("proxy server failed")

		close(done)
	}()

	<-done
}
