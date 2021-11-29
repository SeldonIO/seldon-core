package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent"
	log "github.com/sirupsen/logrus"
)

var (
	serverName       string
	replicaIdx       uint
	schedulerHost    string
	schedulerPort    int
	rcloneHost       string
	rclonePort       int
	inferenceHost    string
	inferencePort    int
	modelRepository  string
	namespace        string
	replicaConfigStr string
	inferenceSvcName string
	configPath       string
)

const (
	EnvMLServerHttpPort = "MLSERVER_HTTP_PORT"
	EnvServerName       = "SELDON_SERVER_NAME"
	EnvServerIdx        = "POD_NAME"
	EnvSchedulerHost    = "SELDON_SCHEDULER_HOST"
	EnvSchedulerPort    = "SELDON_SCHEDULER_PORT"
	EnvReplicaConfig    = "SELDON_REPLICA_CONFIG"

	FlagSchedulerHost = "scheduler-host"
	FlagSchedulerPort = "scheduler-port"
	FlagServerName    = "server-name"
	FlagServerIdx     = "server-idx"
	FlagInferencePort = "inference-port"
	FlagReplicaConfig = "replica-config"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&serverName, FlagServerName, "mlserver", "Server name")
	flag.UintVar(&replicaIdx, "server-idx", 0, "server index")
	flag.StringVar(&schedulerHost, FlagSchedulerHost, "0.0.0.0", "Scheduler host")
	flag.IntVar(&schedulerPort, FlagSchedulerPort, 9005, "Scheduler port")
	flag.StringVar(&rcloneHost, "rclone-host", "0.0.0.0", "RClone host")
	flag.IntVar(&rclonePort, "rclone-port", 5572, "RClone server port")
	flag.StringVar(&inferenceHost, "inference-host", "0.0.0.0", "Inference server host")
	flag.IntVar(&inferencePort, FlagInferencePort, 8080, "Inference server port")
	flag.StringVar(&modelRepository, "model-repository", "/mnt/models", "Model repository folder")
	flag.StringVar(&replicaConfigStr, FlagReplicaConfig, "", "Replica Json Config")
	flag.StringVar(&namespace, "namespace", "default", "Namespace")
	flag.StringVar(&configPath, "config-path", "", "Path to folder with configuration files. Will assume agent.yaml or agent.json in this folder")
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func updateFlagsFromEnv() {
	if !isFlagPassed(FlagInferencePort) {
		port := os.Getenv(EnvMLServerHttpPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagInferencePort, EnvMLServerHttpPort, port)
			var err error
			inferencePort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvMLServerHttpPort, port)
			}
		}
	}
	if !isFlagPassed(FlagServerName) {
		val := os.Getenv(EnvServerName)
		if val != "" {
			log.Infof("Got %s from %s setting to %s", FlagServerName, EnvServerName, val)
			serverName = val
		}
	}
	if !isFlagPassed(FlagServerIdx) {
		podName := os.Getenv(EnvServerIdx)
		if podName != "" {
			lastDashIdx := strings.LastIndex(podName, "-")
			if lastDashIdx == -1 {
				log.Info("Can't decypher pod name to find last dash and index")
				return
			}
			val := podName[lastDashIdx+1:]
			var err error
			idxAsInt, err := strconv.Atoi(val)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvServerIdx, val)
			}
			replicaIdx = uint(idxAsInt)
			log.Infof("Got %s from %s with value %s setting with %d", FlagServerIdx, EnvServerIdx, podName, replicaIdx)
		}
	}
	if !isFlagPassed(FlagSchedulerHost) {
		val := os.Getenv(EnvSchedulerHost)
		if val != "" {
			log.Infof("Got %s from %s setting to %s", FlagSchedulerHost, EnvSchedulerHost, val)
			schedulerHost = val
		}
	}
	if !isFlagPassed(FlagSchedulerPort) {
		port := os.Getenv(EnvSchedulerPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagSchedulerPort, EnvSchedulerPort, port)
			var err error
			schedulerPort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvSchedulerPort, port)
			}
		}
	}
	if !isFlagPassed(FlagReplicaConfig) {
		val := os.Getenv(EnvReplicaConfig)
		if val != "" {
			log.Infof("Got %s from %s setting to %s", FlagReplicaConfig, EnvReplicaConfig, val)
			replicaConfigStr = val
		}
	}
}

func getNamespace() string {
	nsBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Warn("Using namespace from command line argument")
		return namespace
	}
	ns := string(nsBytes)
	log.Info("Namespace is ", ns)
	return ns
}

func setInferenceSvcName() {
	podName := os.Getenv(EnvServerIdx)
	if podName != "" {
		inferenceSvcName = podName
	} else {
		inferenceSvcName = inferenceHost
	}
	log.Infof("Setting inference svc name to %s", inferenceSvcName)
}

func main() {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	flag.Parse()
	updateFlagsFromEnv()
	setInferenceSvcName()

	replicaConfig, err := agent.ParseReplicConfig(replicaConfigStr)
	if err != nil {
		log.Fatalf("Failed to parse replica config %s", replicaConfigStr)
	}

	agentConfigHandler, err := agent.NewAgentConfigHandler(configPath, namespace, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create agent config handler")
	}
	rcloneClient := agent.NewRCloneClient(rcloneHost, rclonePort, modelRepository, logger)
	v2Client := agent.NewV2Client(inferenceHost, inferencePort, logger)
	client, err := agent.NewClient(serverName, uint32(replicaIdx), schedulerHost, schedulerPort, logger, rcloneClient, v2Client, replicaConfig, inferenceSvcName, namespace, agentConfigHandler)
	if err != nil {
		logger.WithError(err).Fatal("Failed to  client")
	}

	err = client.Start()
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialise client")
	}

}
