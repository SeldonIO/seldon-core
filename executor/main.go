package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"net/http"
	"os"
	"os/signal"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"syscall"
	"time"
)

var (
	configPath    = flag.String("config", "", "Path to kubconfig")
	sdepName      = flag.String("sdep", "", "Seldon deployment name")
	namespace     = flag.String("namespace", "", "Namespace")
	predictorName = flag.String("predictor", "", "Name of the predictor inside the SeldonDeployment")
	port          = flag.Int("port", 8080, "Executor port")
	wait          = flag.Duration("graceful-timeout", time.Second*15, "Graceful shutdown secs")
)

func getPredictorFromEnv() (*v1alpha2.PredictorSpec, error) {
	b64Predictor := os.Getenv("ENGINE_PREDICTOR")
	if b64Predictor != "" {
		bytes, err := base64.StdEncoding.DecodeString(b64Predictor)
		if err != nil {
			return nil, err
		}
		predictor := v1alpha2.PredictorSpec{}
		if err := json.Unmarshal(bytes, &predictor); err != nil {
			return nil, err
		} else {
			return &predictor, nil
		}
	} else {
		return nil, nil
	}
}

func main() {
	flag.Parse()

	if *sdepName == "" {
		log.Error("Seldon deployment name must be provided")
		os.Exit(-1)
	}

	if *namespace == "" {
		log.Error("Namespace must be provied")
		os.Exit(-1)
	}

	if *predictorName == "" {
		log.Error("Predictor must be provied")
		os.Exit(-1)
	}

	logf.SetLogger(logf.ZapLogger(false))
	logger := logf.Log.WithName("entrypoint")

	logger.Info("Trying to get predictor from Env")
	predictor, err := getPredictorFromEnv()
	if err != nil {
		logger.Error(err, "Failed to get predictor from Env")
		panic(err)
	} else if predictor == nil {
		logger.Info("Trying to get predictor from API")
		seldonDeploymentClient := client.NewSeldonDeploymentClient(configPath)
		predictor, err = seldonDeploymentClient.GetPredcitor(*sdepName, *namespace, *predictorName)
		if err != nil {
			logger.Error(err, "Failed to find predictor", "name", predictor)
			panic(err)
		}
	}

	// Create REST client
	seldonRest := rest.NewSeldonRestApi(predictor)
	seldonRest.Initialise()

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	logger.Info("Listening", "Address", address)

	srv := &http.Server{
		Handler: seldonRest.Router,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error(err, "Server error")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) and SIGTERM
	// SIGKILL, SIGQUIT will not be caught.
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), *wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logger.Info("shutting down")
	os.Exit(0)
}
