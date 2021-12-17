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
