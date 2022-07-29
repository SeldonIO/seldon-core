package pipeline

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/pkg/metrics"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

const (
	ResourceNameVariable = "model"
)

type GatewayHttpServer struct {
	port    int
	router  *mux.Router
	server  *http.Server
	logger  log.FieldLogger
	ssl     *TLSDetails
	gateway PipelineInferer
	metrics metrics.PipelineMetricsHandler
}

type TLSDetails struct {
	CertMountPath string
	CertFilename  string
	KeyFilename   string
}

func NewGatewayHttpServer(port int, logger log.FieldLogger, ssl *TLSDetails, gateway PipelineInferer, metrics metrics.PipelineMetricsHandler) *GatewayHttpServer {
	return &GatewayHttpServer{
		port:    port,
		router:  mux.NewRouter(),
		logger:  logger.WithField("source", "GatewayHttpServer"),
		ssl:     ssl,
		gateway: gateway,
		metrics: metrics,
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
	if g.ssl != nil && len(g.ssl.CertMountPath) > 0 {
		g.logger.Infof("Creating TLS listener on port %d", g.port)
		certPath := path.Join(g.ssl.CertMountPath, g.ssl.CertFilename)
		keyPath := path.Join(g.ssl.CertMountPath, g.ssl.KeyFilename)
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			log.Fatalf("Error certificate could not be found: %v", err)
		}
		lis, err = tls.Listen("tcp", fmt.Sprintf(":%d", g.port), &tls.Config{Certificates: []tls.Certificate{cert}})
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
		"/v2/models/{" + ResourceNameVariable + "}/infer").HandlerFunc(
		g.metrics.AddPipelineHistogramMetricsHandler(g.inferModel))
	g.router.NewRoute().Path(
		"/v2/pipelines/{" + ResourceNameVariable + "}/infer").HandlerFunc(
		g.metrics.AddPipelineHistogramMetricsHandler(g.inferModel))
}

func (g *GatewayHttpServer) infer(w http.ResponseWriter, req *http.Request, resourceName string, isModel bool) {
	logger := g.logger.WithField("func", "infer")
	startTime := time.Now()
	data, err := ioutil.ReadAll(req.Body)
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
	kafkaRequest, err := g.gateway.Infer(req.Context(), resourceName, isModel, dataProto, convertHttpHeadersToKafkaHeaders(req.Header))
	elapsedTime := time.Since(startTime).Seconds()
	for k, vals := range convertKafkaHeadersToHttpHeaders(kafkaRequest.headers) {
		for _, val := range vals {
			w.Header().Set(k, val)
		}
	}
	w.Header().Set(RequestIdHeader, kafkaRequest.key)
	if err != nil {
		logger.WithError(err).Error("Failed to call infer")
		w.WriteHeader(http.StatusInternalServerError)
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

func (g *GatewayHttpServer) inferModel(w http.ResponseWriter, req *http.Request) {
	logger := g.logger.WithField("func", "inferModel")
	g.logger.Debugf("Seldon model header %s and seldon internal model header %s", req.Header.Get(resources.SeldonModelHeader), req.Header.Get(resources.SeldonInternalModelHeader))
	header := req.Header.Get(resources.SeldonInternalModelHeader) // Seldon internal header takes precedence
	if header == "" {                                             // If we can't find an internal header then look for public one
		header = req.Header.Get(resources.SeldonModelHeader)
	}
	resourceName, isModel, err := createResourceNameFromHeader(header)
	if err != nil {
		logger.WithError(err).Errorf("Failed to create resource name from %s", header)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g.infer(w, req, resourceName, isModel)
}

func (g *GatewayHttpServer) inferPipeline(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	resourceName := vars[ResourceNameVariable]
	g.infer(w, req, resourceName, false)
}
