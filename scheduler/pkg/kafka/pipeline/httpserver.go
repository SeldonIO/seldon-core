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
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

const (
	ModelHttpPathVariable = "model"
)

type GatewayHttpServer struct {
	port    int
	router  *mux.Router
	server  *http.Server
	logger  log.FieldLogger
	ssl     *TLSDetails
	gateway PipelineInferer
}

type TLSDetails struct {
	CertMountPath string
	CertFilename  string
	KeyFilename   string
}

func NewGatewayHttpServer(port int, logger log.FieldLogger, ssl *TLSDetails, gateway PipelineInferer) *GatewayHttpServer {
	return &GatewayHttpServer{
		port:    port,
		router:  mux.NewRouter(),
		logger:  logger.WithField("source", "GatewayHttpServer"),
		ssl:     ssl,
		gateway: gateway,
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
	g.router.NewRoute().Path("/v2/models/{" + ModelHttpPathVariable + "}/infer").HandlerFunc(g.infer)
}

func (g *GatewayHttpServer) infer(w http.ResponseWriter, req *http.Request) {
	logger := g.logger.WithField("func", "infer")
	header := req.Header.Get(resources.SeldonModelHeader)
	resourceName, isModel, err := createResourceNameFromHeader(header)
	if err != nil {
		logger.WithError(err).Errorf("Failed to create resource name from %s", header)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
	res, err := g.gateway.Infer(req.Context(), resourceName, isModel, dataProto)
	if err != nil {
		logger.WithError(err).Error("Failed to call infer")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resJson, err := ConvertV2ResponseBytesToJson(res)
		if err != nil {
			logger.WithError(err).Errorf("Failed to convert v2 response to json for resource %s", resourceName)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(resJson)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
