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

package main

import (
	"os"
	"os/signal"
	"syscall"

	proxy "github.com/seldonio/seldon-core/scheduler/pkg/proxy"
	"github.com/sirupsen/logrus"
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
