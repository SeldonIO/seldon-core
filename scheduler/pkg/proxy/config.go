package proxy

import (
	"flag"

	"github.com/sirupsen/logrus"
)

type Config struct {
	LogConfig
	ProxyConfig
	AgentConfig
	ChainerConfig
}

const (
	LogDefaultLevel = logrus.InfoLevel
)

type LogConfig struct {
	LogLevel logrus.Level
}

const (
	ProxyDefaultListenPort      = 8100
	ProxyDefaultEventBufferSize = 100
)

type ProxyConfig struct {
	ProxyListenPort      uint
	ProxyEventBufferSize uint
}

const (
	AgentDefaultListenPort = 9005
)

type AgentConfig struct {
	AgentListenPort uint
}

const (
	ChainerDefaultListenPort = 10101
)

type ChainerConfig struct {
	ChainerListenPort uint
}

func GetConfig() *Config {
	return ParseArgs()
}

func ParseArgs() *Config {
	logLevel := flag.String("level", "info", "Log level, e.g. info warn error debug")
	agentListenPort := flag.Uint("agentPort", AgentDefaultListenPort, "Port on which to listen for agent connections")
	chainerListenPort := flag.Uint("chainerPort", ChainerDefaultListenPort, "Port on which to listen for data-flow chainer connections")
	proxyListenPort := flag.Uint("proxyPort", ProxyDefaultListenPort, "Port on which to listen for model requests")
	proxyBufferSize := flag.Uint("eventBufferSize", ProxyDefaultEventBufferSize, "Number of pending model requests permitted")

	flag.Parse()

	l, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		l = LogDefaultLevel
	}

	c := Config{}
	c.LogLevel = l
	c.AgentListenPort = *agentListenPort
	c.ChainerListenPort = *chainerListenPort
	c.ProxyListenPort = *proxyListenPort
	c.ProxyEventBufferSize = *proxyBufferSize

	return &c
}
