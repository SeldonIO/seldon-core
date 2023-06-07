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

package pipeline

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/status"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	ResourceNameVariable = "model"
	v2ModelPathPrefix    = "/v2/models/"
	v2PipelinePathPrefix = "/v2/pipelines/"
)

type GatewayHttpServer struct {
	port                 int
	router               *mux.Router
	server               *http.Server
	logger               log.FieldLogger
	gateway              PipelineInferer
	metrics              metrics.PipelineMetricsHandler
	tlsOptions           *util.TLSOptions
	pipelineReadyChecker status.PipelineReadyChecker
}

type TLSDetails struct {
	CertMountPath string
	CertFilename  string
	KeyFilename   string
}

func NewGatewayHttpServer(port int, logger log.FieldLogger,
	gateway PipelineInferer,
	metrics metrics.PipelineMetricsHandler,
	tlsOptions *util.TLSOptions,
	pipelineReadyChecker status.PipelineReadyChecker) *GatewayHttpServer {
	return &GatewayHttpServer{
		port:                 port,
		router:               mux.NewRouter(),
		logger:               logger.WithField("source", "GatewayHttpServer"),
		gateway:              gateway,
		metrics:              metrics,
		tlsOptions:           tlsOptions,
		pipelineReadyChecker: pipelineReadyChecker,
	}
}

func (g *GatewayHttpServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel()
	return g.server.Shutdown(ctx)
}

func (g *GatewayHttpServer) Start() error {
	logger := g.logger.WithField("func", "Start")
	logger.Infof("Starting http server on port %d", g.port)
	g.setupRoutes()
	g.server = &http.Server{
		Handler:     g.router,
		IdleTimeout: 65 * time.Second,
	}
	lis := g.createListener()
	return g.server.Serve(lis)
}

func (g *GatewayHttpServer) createListener() net.Listener {
	// Create a listener at the desired port.
	var lis net.Listener
	var err error
	if g.tlsOptions.TLS {
		g.logger.Infof("Creating TLS listener on port %d", g.port)

		lis, err = tls.Listen("tcp", fmt.Sprintf(":%d", g.port), g.tlsOptions.Cert.CreateServerTLSConfig())
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	} else {
		g.logger.Infof("Creating non-TLS listener port %d", g.port)
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", g.port))
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	}
	return lis
}

func (g *GatewayHttpServer) setupRoutes() {
	g.router.Use(mux.CORSMethodMiddleware(g.router))
	g.router.Use(otelmux.Middleware("pipelinegateway"))
	g.router.NewRoute().Path(
		v2ModelPathPrefix + "{" + ResourceNameVariable + "}/infer").HandlerFunc(g.inferModel)
	g.router.NewRoute().Path(
		v2PipelinePathPrefix + "{" + ResourceNameVariable + "}/infer").HandlerFunc(g.inferPipeline)
	g.router.NewRoute().Path(
		v2ModelPathPrefix + "{" + ResourceNameVariable + "}/ready").HandlerFunc(g.pipelineReadyFromModelPath)
	g.router.NewRoute().Path(
		v2PipelinePathPrefix + "{" + ResourceNameVariable + "}/ready").HandlerFunc(g.pipelineReadyFromPipelinePath)
}

// Get or create a request ID
func (g *GatewayHttpServer) getRequestId(req *http.Request) string {
	requestIds := req.Header[util.RequestIdHeaderCanonical]
	var requestId string
	if len(requestIds) > 0 {
		requestId = requestIds[0]
	} else {
		g.logger.Warning("Failed to find request ID - will generate one")
		requestId = util.CreateRequestId()
	}
	return requestId
}

func (g *GatewayHttpServer) infer(w http.ResponseWriter, req *http.Request, resourceName string, isModel bool) {
	logger := g.logger.WithField("func", "infer")
	startTime := time.Now()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dataProto, err := ConvertRequestToV2Bytes(data, "", "")
	if err != nil {
		logger.WithError(err).Errorf("Failed to convert bytes to v2 request for resource %s", resourceName)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	kafkaRequest, err := g.gateway.Infer(req.Context(), resourceName, isModel, dataProto, convertHttpHeadersToKafkaHeaders(req.Header), g.getRequestId(req))
	elapsedTime := time.Since(startTime).Seconds()
	if kafkaRequest != nil {
		for k, vals := range convertKafkaHeadersToHttpHeaders(kafkaRequest.headers) {
			for _, val := range vals {
				w.Header().Add(k, val)
			}
		}
	}
	w.Header().Set(util.RequestIdHeader, kafkaRequest.key)
	if err != nil {
		logger.WithError(err).Error("Failed to call infer")
		w.WriteHeader(http.StatusInternalServerError)
	} else if kafkaRequest.isError {
		logger.Error(string(kafkaRequest.response))
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(createResponseErrorPayload(kafkaRequest.errorModel, kafkaRequest.response))
		if err != nil {
			logger.WithError(err).Error("Failed to write error payload")
		}
	} else {
		resJson, err := ConvertV2ResponseBytesToJson(kafkaRequest.response)
		if err != nil {
			logger.WithError(err).Errorf("Failed to convert v2 response to json for resource %s", resourceName)
			go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeRest, elapsedTime, metrics.HttpCodeToString(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(resJson)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeRest, elapsedTime, metrics.HttpCodeToString(http.StatusOK))
		}
	}
}

func getResourceFromHeaders(req *http.Request, logger log.FieldLogger) (string, bool, error) {
	modelHeader := req.Header.Get(resources.SeldonModelHeader)
	// may have multiple header values due to shadow/mirror processing
	modelInternalHeader := req.Header.Values(resources.SeldonInternalModelHeader)
	logger.Debugf("Seldon model header %s and seldon internal model header %s", modelHeader, modelInternalHeader)
	if len(modelInternalHeader) > 0 {
		return createResourceNameFromHeader(modelInternalHeader[len(modelInternalHeader)-1]) // get last header if multiple
	} else {
		return createResourceNameFromHeader(modelHeader)
	}
}

func (g *GatewayHttpServer) inferModel(w http.ResponseWriter, req *http.Request) {
	logger := g.logger.WithField("func", "inferModel")
	resourceName, isModel, err := getResourceFromHeaders(req, logger)
	if err != nil {
		logger.WithError(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g.infer(w, req, resourceName, isModel)
}

func (g *GatewayHttpServer) inferPipeline(w http.ResponseWriter, req *http.Request) {
	logger := g.logger.WithField("func", "inferPipeline")
	resourceName, isModel, err := getResourceFromHeaders(req, logger)
	if err != nil {
		logger.Error("No header found for pipeline identification")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g.infer(w, req, resourceName, isModel)
}

func (g *GatewayHttpServer) pipelineReady(w http.ResponseWriter, req *http.Request, resourceName string) {
	logger := g.logger.WithField("func", "pipelineReady")
	ready, err := g.pipelineReadyChecker.CheckPipelineReady(req.Context(), resourceName, g.getRequestId(req))
	if err != nil {
		if errors.Is(err, status.PipelineNotFoundErr) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.WithError(err).Errorf("Failed to get pipeline readines for pipeline %s", resourceName)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		if ready {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusFailedDependency)
		}
	}
}

func (g *GatewayHttpServer) pipelineReadyFromPipelinePath(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	resourceName := vars[ResourceNameVariable]
	g.pipelineReady(w, req, resourceName)

}

func (g *GatewayHttpServer) pipelineReadyFromModelPath(w http.ResponseWriter, req *http.Request) {
	logger := g.logger.WithField("func", "inferModel")
	resourceName, isModel, err := getResourceFromHeaders(req, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to create resource name from header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if isModel {
		logger.Errorf("Model ready call to pipeline gateway. Will ignore")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g.pipelineReady(w, req, resourceName)
}
