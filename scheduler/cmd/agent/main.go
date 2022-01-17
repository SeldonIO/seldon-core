package main

import (
	"flag"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/repository/mlserver"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/repository/triton"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/rclone"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/repository"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
	"k8s.io/client-go/kubernetes"

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
	agentFolder      string
	namespace        string
	replicaConfigStr string
	inferenceSvcName string
	configPath       string
	logLevel         string
	serverType       string
	serverTypes      = [...]string{"mlserver", "triton"}
)

const (
	EnvServerHttpPort = "SELDON_SERVER_HTTP_PORT"
	EnvServerName     = "SELDON_SERVER_NAME"
	EnvServerIdx      = "POD_NAME"
	EnvSchedulerHost  = "SELDON_SCHEDULER_HOST"
	EnvSchedulerPort  = "SELDON_SCHEDULER_PORT"
	EnvReplicaConfig  = "SELDON_REPLICA_CONFIG"
	EnvLogLevel       = "SELDON_LOG_LEVEL"
	EnvServerType     = "SELDON_SERVER_TYPE"

	FlagSchedulerHost = "scheduler-host"
	FlagSchedulerPort = "scheduler-port"
	FlagServerName    = "server-name"
	FlagServerIdx     = "server-idx"
	FlagInferencePort = "inference-port"
	FlagReplicaConfig = "replica-config"
	FlagLogLevel      = "log-level"
	FlagServerType    = "server-type"
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
	flag.StringVar(&agentFolder, "agent-folder", "/mnt/agent", "Model repository folder")
	flag.StringVar(&replicaConfigStr, FlagReplicaConfig, "", "Replica Json Config")
	flag.StringVar(&namespace, "namespace", "", "Namespace")
	flag.StringVar(&configPath, "config-path", "/mnt/config", "Path to folder with configuration files. Will assume agent.yaml or agent.json in this folder")
	flag.StringVar(&logLevel, FlagLogLevel, "debug", "Log level - examples: debug, info, error")
	flag.StringVar(&serverType, FlagServerType, serverTypes[0], "server type. Default mlserver")
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
		port := os.Getenv(EnvServerHttpPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagInferencePort, EnvServerHttpPort, port)
			var err error
			inferencePort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvServerHttpPort, port)
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
	if !isFlagPassed(FlagLogLevel) {
		val := os.Getenv(EnvLogLevel)
		if val != "" {
			log.Infof("Got %s from %s setting to %s", FlagLogLevel, EnvLogLevel, val)
			logLevel = val
		}
	}
	if !isFlagPassed(FlagServerType) {
		val := os.Getenv(EnvServerType)
		if val != "" {
			log.Infof("Got %s from %s setting to %s", FlagServerType, EnvServerType, val)
			serverType = val
		}
	}
}

func runningInsideK8s() bool {
	return namespace != ""
}

func updateNamespace() {
	nsBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Warn("Using namespace from command line argument")
	} else {
		namespace = string(nsBytes)
	}
	if runningInsideK8s() {
		log.Info("Running inside k8s. Namespace is ", namespace)
	}
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

func updateFlags() {
	updateFlagsFromEnv()
	setInferenceSvcName()
	updateNamespace()
}

func makeDirs() (string, string, error) {
	modelRepositoryDir := filepath.Join(agentFolder, "models")
	rcloneRepositoryDir := filepath.Join(agentFolder, "rclone")
	err := os.MkdirAll(modelRepositoryDir, fs.ModePerm)
	if err != nil {
		return modelRepositoryDir, rcloneRepositoryDir, err
	}
	err = os.MkdirAll(rcloneRepositoryDir, fs.ModePerm)
	return modelRepositoryDir, rcloneRepositoryDir, err
}

func getRepositoryHandler(logger log.FieldLogger) repository.ModelRepositoryHandler {
	switch serverType {
	case "mlserver":
		logger.Infof("Creating MLServer repository handler")
		return mlserver.NewMLServerRepositoryHandler(logger)
	case "triton":
		logger.Infof("Creating Triton repository handler")
		return triton.NewTritonRepositoryHandler(logger)
	default:
		logger.Infof("Using default as no server type requested - creating MLServer repository handler")
		return mlserver.NewMLServerRepositoryHandler(logger)
	}
}

func main() {
	logger := log.New()
	flag.Parse()
	updateFlags()
	logIntLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", logLevel)
	}
	logger.Infof("Setting log level to %s", logLevel)
	logger.SetLevel(logIntLevel)

	// Make required folders
	//TODO handle via initContainer?
	modelRepositoryDir, rcloneRepositoryDir, err := makeDirs()
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create required folders %s and %s", modelRepositoryDir, rcloneRepositoryDir)
	}
	log.Infof("Model repository dir %s, Rclone repository dir %s ", modelRepositoryDir, rcloneRepositoryDir)

	done := make(chan bool, 1)

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		logger.Info("shutting down due to SIGTERM or SIGINT")
		close(done)
	}()

	replicaConfig, err := agent.ParseReplicaConfig(replicaConfigStr)
	if err != nil {
		log.Fatalf("Failed to parse replica config %s", replicaConfigStr)
	}

	var clientset kubernetes.Interface
	if runningInsideK8s() {
		clientset, err = k8s.CreateClientset()
		if err != nil { //TODO change to Error from Fatal?
			logger.WithError(err).Fatal("Failed to create kubernetes clientset")
		}
	}
	// Start Agent configuration handler
	agentConfigHandler, err := config.NewAgentConfigHandler(configPath, namespace, logger, clientset)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create agent config handler")
	}
	defer func() {
		_ = agentConfigHandler.Close()
		logger.Info("Closed agent handler")
	}()

	// Create Rclone client and start configuration listener

	rcloneClient := rclone.NewRCloneClient(rcloneHost, rclonePort, rcloneRepositoryDir, logger, namespace)
	err = rcloneClient.StartConfigListener(agentConfigHandler)
	if err != nil {
		logger.WithError(err).Error("Failed to initialise rclone config listener")
		close(done)
	}

	// Create Model Repository
	modelRepository := repository.NewModelRepository(logger, rcloneClient, modelRepositoryDir, getRepositoryHandler(logger))
	// Create V2 Protocol Handler
	v2Client := agent.NewV2Client(inferenceHost, inferencePort, logger)
	// Create Agent
	client := agent.NewClient(serverName, uint32(replicaIdx), schedulerHost, schedulerPort, logger, modelRepository, v2Client, replicaConfig, inferenceSvcName, namespace)

	// Start client grpc server
	go func() {
		err = client.Start()
		if err != nil {
			logger.WithError(err).Error("Failed to initialise client")
		}
		close(done)
	}()

	// Wait for completion
	<-done
}
