package main

import (
	"errors"
	"flag"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/metrics"

	agent2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"

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
	agentHost            string
	serverName           string
	replicaIdx           uint
	schedulerHost        string
	schedulerPort        int
	rcloneHost           string
	rclonePort           int
	inferenceHost        string
	inferenceHttpPort    int
	inferenceGrpcPort    int
	reverseProxyHttpPort int
	reverseProxyGrpcPort int
	debugGrpcPort        int
	metricsPort          int
	agentFolder          string
	namespace            string
	replicaConfigStr     string
	inferenceSvcName     string
	configPath           string
	logLevel             string
	serverType           string
	memoryBytes          int
	memoryBytes64        uint64
	capabilitiesList     string
	capabilities         []string
	overCommit           bool
	serverTypes          = [...]string{"mlserver", "triton"}
)

const (
	EnvServerHttpPort       = "SELDON_SERVER_HTTP_PORT"
	EnvServerGrpcPort       = "SELDON_SERVER_GRPC_PORT"
	EnvReverseProxyHttpPort = "SELDON_REVERSE_PROXY_HTTP_PORT"
	EnvReverseProxyGrpcPort = "SELDON_REVERSE_PROXY_GRPC_PORT"
	EnvDebugGrpcPort        = "SELDON_DEBUG_GRPC_PORT"
	EnvPodName              = "POD_NAME"
	EnvSchedulerHost        = "SELDON_SCHEDULER_HOST"
	EnvSchedulerPort        = "SELDON_SCHEDULER_PORT"
	EnvReplicaConfig        = "SELDON_REPLICA_CONFIG"
	EnvLogLevel             = "SELDON_LOG_LEVEL"
	EnvServerType           = "SELDON_SERVER_TYPE"
	EnvMemoryRequest        = "MEMORY_REQUEST"
	EnvCapabilities         = "SELDON_SERVER_CAPABILITIES"
	EnvOvercommit           = "SELDON_OVERCOMMIT"
	EnvMetricsPort          = "SELDON_METRICS_PORT"

	FlagSchedulerHost        = "scheduler-host"
	FlagSchedulerPort        = "scheduler-port"
	FlagServerName           = "server-name"
	FlagServerIdx            = "server-idx"
	FlagInferenceHttpPort    = "inference-http-port"
	FlagInferenceGrpcPort    = "inference-grpc-port"
	FlagReverseProxyHttpPort = "reverse-proxy-http-port"
	FlagReverseProxyGrpcPort = "reverse-proxy-grpc-port"
	FlagDebugGrpcPort        = "debug-grpc-port"
	FlagReplicaConfig        = "replica-config"
	FlagLogLevel             = "log-level"
	FlagServerType           = "server-type"
	FlagMemoryBytes          = "memory-bytes"
	FlagCapabilities         = "capabilities"
	FlagOverCommit           = "overcommit"
	FlagMetricsPort          = "metrics-port"

	DefaultInferenceHttpPort = 8080
	DefaultInferenceGrpcPort = 9500
	DefaultRclonePort        = 5572
	DefaultSchedulerPort     = 9005
	DefaultMetricsPort       = 9006
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&agentHost, "agent-host", "0.0.0.0", "Agent hostname")
	flag.StringVar(&serverName, FlagServerName, "mlserver", "Server name")
	flag.UintVar(&replicaIdx, "server-idx", 0, "Server index")
	flag.StringVar(&schedulerHost, FlagSchedulerHost, "0.0.0.0", "Scheduler host")
	flag.IntVar(&schedulerPort, FlagSchedulerPort, DefaultSchedulerPort, "Scheduler port")
	flag.StringVar(&rcloneHost, "rclone-host", "0.0.0.0", "RClone host")
	flag.IntVar(&rclonePort, "rclone-port", DefaultRclonePort, "RClone server port")
	flag.StringVar(&inferenceHost, "inference-host", "0.0.0.0", "Inference server host")
	flag.IntVar(&inferenceHttpPort, FlagInferenceHttpPort, DefaultInferenceHttpPort, "Inference server http port")
	flag.IntVar(&inferenceGrpcPort, FlagInferenceGrpcPort, DefaultInferenceGrpcPort, "Inference server grpc port")
	flag.IntVar(&reverseProxyHttpPort, FlagReverseProxyHttpPort, agent.ReverseProxyHTTPPort, "Reverse proxy http port")
	flag.IntVar(&reverseProxyGrpcPort, FlagReverseProxyGrpcPort, agent.ReverseGRPCProxyPort, "Reverse proxy grpc port")
	flag.IntVar(&debugGrpcPort, FlagDebugGrpcPort, agent.GRPCDebugServicePort, "Debug grpc port")
	flag.StringVar(&agentFolder, "agent-folder", "/mnt/agent", "Model repository folder")
	flag.StringVar(&replicaConfigStr, FlagReplicaConfig, "", "Replica Json Config")
	flag.StringVar(&namespace, "namespace", "", "Namespace")
	flag.StringVar(&configPath, "config-path", "/mnt/config", "Path to folder with configuration files. Will assume agent.yaml or agent.json in this folder")
	flag.StringVar(&serverType, FlagServerType, serverTypes[0], "Server type. Default mlserver")
	flag.IntVar(&memoryBytes, FlagMemoryBytes, 1000000, "Memory available for server")
	flag.StringVar(&capabilitiesList, FlagCapabilities, "sklearn,xgboost", "Server capabilities")
	flag.BoolVar(&overCommit, FlagOverCommit, true, "Overcommit memory")
	flag.StringVar(&logLevel, FlagLogLevel, "debug", "Log level - examples: debug, info, error")
	flag.IntVar(&metricsPort, FlagMetricsPort, DefaultMetricsPort, "Metrics Port")
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

func setServerNameAndIdxFromPodName() {
	log.Infof("Trying to set server name and replica index from pod name")
	podName := os.Getenv(EnvPodName)
	if podName != "" {
		lastDashIdx := strings.LastIndex(podName, "-")
		if lastDashIdx == -1 {
			log.Infof("Can't decypher pod name to find last dash and index. %s", podName)
		} else {
			serverIdxStr := podName[lastDashIdx+1:]
			var err error
			serverIdx, err := strconv.Atoi(serverIdxStr)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse to integer %s with value %s", EnvPodName, serverIdxStr)
			} else {
				replicaIdx = uint(serverIdx)
				serverName = podName[0:lastDashIdx]
				log.Infof("Got server name and index from %s with value %s. Server name:%s Replica Idx:%d", EnvPodName, podName, serverName, replicaIdx)
			}
		}
	}
}

func updateFlagsFromEnv() {
	if !isFlagPassed(FlagOverCommit) {
		var err error
		overCommitStr := os.Getenv(EnvOvercommit)
		if overCommitStr != "" {
			overCommit, err = strconv.ParseBool(overCommitStr)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s for overcommit", EnvOvercommit, overCommitStr)
			}
			log.Infof("Setting overcommit from env %s with value %s to %v", EnvOvercommit, overCommitStr, overCommit)
		}
	}
	if !isFlagPassed(FlagCapabilities) {
		capabilitiesList := os.Getenv(EnvCapabilities)
		log.Infof("Updating capabilities from env %s with value %s", EnvCapabilities, capabilitiesList)
		capabilities = strings.Split(capabilitiesList, ",")
	} else {
		capabilities = strings.Split(capabilitiesList, ",")
	}
	log.Infof("Server Capabilities %v", capabilities)
	if !isFlagPassed(FlagMemoryBytes) {
		memoryRequests := os.Getenv(EnvMemoryRequest)
		var err error
		if memoryRequests != "" {
			memoryBytes64, err = strconv.ParseUint(memoryRequests, 10, 64)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvMemoryRequest, memoryRequests)
			}
		} else {
			memoryBytes64 = uint64(memoryBytes)
		}
	} else {
		memoryBytes64 = uint64(memoryBytes)
	}
	if !isFlagPassed(FlagInferenceHttpPort) {
		port := os.Getenv(EnvServerHttpPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagInferenceHttpPort, EnvServerHttpPort, port)
			var err error
			inferenceHttpPort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvServerHttpPort, port)
			}
		}
	}
	if !isFlagPassed(FlagInferenceGrpcPort) {
		port := os.Getenv(EnvServerGrpcPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagInferenceGrpcPort, EnvServerGrpcPort, port)
			var err error
			inferenceGrpcPort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvServerGrpcPort, port)
			}
		}
	}
	if !isFlagPassed(FlagReverseProxyHttpPort) {
		port := os.Getenv(EnvReverseProxyHttpPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagReverseProxyHttpPort, EnvReverseProxyHttpPort, port)
			var err error
			reverseProxyHttpPort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvReverseProxyHttpPort, port)
			}
		}
	}
	if !isFlagPassed(FlagReverseProxyGrpcPort) {
		port := os.Getenv(EnvReverseProxyGrpcPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagReverseProxyGrpcPort, EnvReverseProxyGrpcPort, port)
			var err error
			reverseProxyGrpcPort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvReverseProxyGrpcPort, port)
			}
		}
	}
	if !isFlagPassed(FlagDebugGrpcPort) {
		port := os.Getenv(EnvDebugGrpcPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagDebugGrpcPort, EnvDebugGrpcPort, port)
			var err error
			debugGrpcPort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvDebugGrpcPort, port)
			}
		}
	}
	if !isFlagPassed(FlagServerName) || !isFlagPassed(FlagServerIdx) {
		setServerNameAndIdxFromPodName()
	} else {
		log.Warnf("Using passed in values for server name and server index. Server name %s server index %d", serverName, replicaIdx)
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
	if !isFlagPassed(FlagMetricsPort) {
		port := os.Getenv(EnvMetricsPort)
		if port != "" {
			log.Infof("Got %s from %s setting to %s", FlagMetricsPort, EnvMetricsPort, port)
			var err error
			metricsPort, err = strconv.Atoi(port)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse %s with value %s", EnvMetricsPort, port)
			}
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
	podName := os.Getenv(EnvPodName)
	if podName != "" {
		inferenceSvcName = podName
	} else {
		inferenceSvcName = agentHost
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

func createReplicaConfig() *agent2.ReplicaConfig {
	var rc *agent2.ReplicaConfig
	if isFlagPassed(FlagReplicaConfig) {
		var err error
		rc, err = agent.ParseReplicaConfig(replicaConfigStr)
		if err != nil {
			log.WithError(err).Fatalf("Failed to parse replica config %s", replicaConfigStr)
		}
		log.Infof("Created replicaConfig from command line")
	} else {
		rc = &agent2.ReplicaConfig{
			InferenceSvc:      inferenceSvcName,
			InferenceHttpPort: int32(inferenceHttpPort),
			InferenceGrpcPort: int32(inferenceGrpcPort),
			MemoryBytes:       memoryBytes64,
			Capabilities:      capabilities,
			OverCommit:        overCommit,
		}
		log.Infof("Created replicaConfig from environment")
	}
	//Point to proxy always in replica config
	rc.InferenceHttpPort = int32(reverseProxyHttpPort)
	rc.InferenceGrpcPort = int32(reverseProxyGrpcPort)
	log.Infof("replicaConfig %+v", rc)
	return rc
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

	// Create Model Repository
	modelRepository := repository.NewModelRepository(logger, rcloneClient, modelRepositoryDir, getRepositoryHandler(logger))
	// Create V2 Protocol Handler
	v2Client := agent.NewV2Client(inferenceHost, inferenceHttpPort, logger)

	promMetrics, err := metrics.NewPrometheusMetrics(serverName, replicaIdx, namespace, logger)
	if err != nil {
		logger.WithError(err).Fatalf("Can't create prometheus metrics")
	}
	go func() {
		err := promMetrics.Start(metricsPort)
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		logger.WithError(err).Error("Can't start metrics server")
		close(done)
	}()
	defer func() { _ = promMetrics.Stop() }()

	rpHTTP := agent.NewReverseHTTPProxy(logger, uint(reverseProxyHttpPort), promMetrics)
	defer func() { _ = rpHTTP.Stop() }()

	rpGRPC := agent.NewReverseGRPCProxy(promMetrics, logger, inferenceHost, uint(inferenceGrpcPort), uint(reverseProxyGrpcPort))
	defer func() { _ = rpGRPC.Stop() }()

	clientDebugService := agent.NewClientDebug(logger, uint(debugGrpcPort))
	defer func() { _ = clientDebugService.Stop() }()

	// Create Agent
	client := agent.NewClient(serverName, uint32(replicaIdx), schedulerHost, schedulerPort, logger, modelRepository, v2Client, createReplicaConfig(), inferenceSvcName, namespace, rpHTTP, rpGRPC, clientDebugService)

	// Wait for required services to be ready
	err = client.WaitReady()
	if err != nil {
		logger.WithError(err).Errorf("Failed to wait for all agent dependent services to be ready")
		close(done)
	}

	// No we are ready start config listener
	err = rcloneClient.StartConfigListener(agentConfigHandler)
	if err != nil {
		logger.WithError(err).Error("Failed to initialise rclone config listener")
		close(done)
	}

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
