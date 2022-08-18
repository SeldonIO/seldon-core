package main

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/seldonio/seldon-core/scheduler/pkg/tracing"

	"github.com/seldonio/seldon-core/scheduler/pkg/metrics"

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

	"github.com/seldonio/seldon-core/scheduler/cmd/agent/cli"
)

func makeDirs() (string, string, error) {
	modelRepositoryDir := filepath.Join(cli.AgentFolder, "models")
	rcloneRepositoryDir := filepath.Join(cli.AgentFolder, "rclone")
	err := os.MkdirAll(modelRepositoryDir, fs.ModePerm)
	if err != nil {
		return modelRepositoryDir, rcloneRepositoryDir, err
	}
	_ = os.Chmod(modelRepositoryDir, fs.ModePerm)
	err = os.MkdirAll(rcloneRepositoryDir, fs.ModePerm)
	if err != nil {
		return modelRepositoryDir, rcloneRepositoryDir, err
	}
	_ = os.Chmod(rcloneRepositoryDir, fs.ModePerm)
	return modelRepositoryDir, rcloneRepositoryDir, nil
}

func getRepositoryHandler(logger log.FieldLogger) repository.ModelRepositoryHandler {
	switch cli.ServerType {
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
	if cli.ReplicaConfigStr != "" {
		var err error
		rc, err = agent.ParseReplicaConfig(cli.ReplicaConfigStr)
		if err != nil {
			log.WithError(err).Fatalf("Failed to parse replica config %s", cli.ReplicaConfigStr)
		}
		log.Infof("Created replicaConfig from command line")
	} else {
		rc = &agent2.ReplicaConfig{
			InferenceSvc:         cli.InferenceSvcName,
			InferenceHttpPort:    int32(cli.InferenceHttpPort),
			InferenceGrpcPort:    int32(cli.InferenceGrpcPort),
			MemoryBytes:          cli.MemoryBytes64,
			Capabilities:         cli.Capabilities,
			OverCommitPercentage: uint32(cli.OverCommitPercentage),
		}
		log.Infof("Created replicaConfig from environment")
	}
	//Point to proxy always in replica config
	rc.InferenceHttpPort = int32(cli.ReverseProxyHttpPort)
	rc.InferenceGrpcPort = int32(cli.ReverseProxyGrpcPort)
	log.Infof("replicaConfig %+v", rc)
	return rc
}

func runningInsideK8s() bool {
	return cli.Namespace != ""
}

func makeSignalHandler(logger *log.Logger, done chan<- bool) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	logger.Info("shutting down due to SIGTERM or SIGINT")
	close(done)
}

func main() {
	logger := log.New()

	cli.UpdateArgs()

	logIntLevel, err := log.ParseLevel(cli.LogLevel)
	if err != nil {
		logger.WithError(err).Fatalf("Failed to set log level %s", cli.LogLevel)
	}
	logger.Infof("Setting log level to %s", cli.LogLevel)
	logger.SetLevel(logIntLevel)

	// Make required folders
	//TODO handle via initContainer?
	modelRepositoryDir, rcloneRepositoryDir, err := makeDirs()
	if err != nil {
		logger.
			WithError(err).
			Fatalf("Failed to create required folders %s and %s", modelRepositoryDir, rcloneRepositoryDir)
	}
	log.Infof("Model repository dir %s, Rclone repository dir %s ", modelRepositoryDir, rcloneRepositoryDir)

	done := make(chan bool, 1)

	go makeSignalHandler(logger, done)

	var clientset kubernetes.Interface
	if runningInsideK8s() {
		clientset, err = k8s.CreateClientset()
		if err != nil { //TODO change to Error from Fatal?
			logger.WithError(err).Fatal("Failed to create kubernetes clientset")
		}
	}

	tracer, err := tracing.NewTraceProvider("seldon-agent", &cli.TracingConfigPath, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to configure otel tracer")
	} else {
		defer tracer.Stop()
	}

	// Start Agent configuration handler
	agentConfigHandler, err := config.NewAgentConfigHandler(cli.ConfigPath, cli.Namespace, logger, clientset)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create agent config handler")
	}
	defer func() {
		_ = agentConfigHandler.Close()
		logger.Info("Closed agent handler")
	}()

	// Create Rclone client
	rcloneClient := rclone.NewRCloneClient(cli.RcloneHost, cli.RclonePort, rcloneRepositoryDir, logger, cli.Namespace)

	// Create Model Repository
	modelRepository := repository.NewModelRepository(
		logger,
		rcloneClient,
		modelRepositoryDir,
		getRepositoryHandler(logger),
		cli.EnvoyHost,
		cli.EnvoyPort,
	)

	// Create V2 Protocol Handler
	v2Client := agent.NewV2Client(cli.InferenceHost, cli.InferenceGrpcPort, logger, true)

	promMetrics, err := metrics.NewPrometheusModelMetrics(cli.ServerName, cli.ReplicaIdx, cli.Namespace, logger)
	if err != nil {
		logger.WithError(err).Fatalf("Can't create prometheus metrics")
	}
	go func() {
		err := promMetrics.Start(cli.MetricsPort)
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		logger.WithError(err).Error("Can't start metrics server")
		close(done)
	}()
	defer func() { _ = promMetrics.Stop() }()

	rpHTTP := agent.NewReverseHTTPProxy(
		logger,
		cli.InferenceHost,
		uint(cli.InferenceHttpPort),
		uint(cli.ReverseProxyHttpPort),
		promMetrics,
	)
	defer func() { _ = rpHTTP.Stop() }()

	rpGRPC := agent.NewReverseGRPCProxy(
		promMetrics,
		logger,
		cli.InferenceHost,
		uint(cli.InferenceGrpcPort),
		uint(cli.ReverseProxyGrpcPort),
	)
	defer func() { _ = rpGRPC.Stop() }()

	clientDebugService := agent.NewClientDebug(logger, uint(cli.DebugGrpcPort))
	defer func() { _ = clientDebugService.Stop() }()

	// Create Agent
	client := agent.NewClient(
		cli.ServerName,
		uint32(cli.ReplicaIdx),
		cli.SchedulerHost,
		cli.SchedulerPort,
		logger,
		modelRepository,
		v2Client,
		createReplicaConfig(),
		cli.InferenceSvcName,
		cli.Namespace,
		rpHTTP,
		rpGRPC,
		clientDebugService,
		promMetrics,
	)

	// Wait for required services to be ready
	err = client.WaitReady()
	if err != nil {
		logger.WithError(err).Errorf("Failed to wait for all agent dependent services to be ready")
		close(done)
	}

	// Now we are ready start config listener
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
