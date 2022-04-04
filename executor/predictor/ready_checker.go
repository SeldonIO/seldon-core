package predictor

import (
	"fmt"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

func Ready(protocol string, node *v1.PredictiveUnit) error {
	switch protocol {
	case api.ProtocolSeldon:
		return ReadyHealth(node, "/api/v1.0/health/status")
	case api.ProtocolTensorflow:
		return ReadyTCP(node)
	case api.ProtocolV2:
		return ReadyHealth(node, "/v2/health/ready")
	default:
		return fmt.Errorf("Unknown protocol for health check: %s", protocol)
	}
}

func ReadyTCP(node *v1.PredictiveUnit) error {
	for _, child := range node.Children {
		err := ReadyTCP(&child)
		if err != nil {
			return err
		}
	}
	if node.Endpoint != nil && node.Endpoint.ServiceHost != "" && node.Endpoint.ServicePort > 0 {
		c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", node.Endpoint.ServiceHost, node.Endpoint.ServicePort))
		if err != nil {
			return err
		} else {
			err = c.Close()
			return nil
		}
	} else {
		return nil
	}
}

func ReadyHealth(node *v1.PredictiveUnit, healthPath string) error {
	for _, child := range node.Children {
		err := ReadyHealth(&child, healthPath)
		if err != nil {
			return err
		}
	}
	if node.Endpoint != nil && node.Endpoint.ServiceHost != "" && node.Endpoint.ServicePort > 0 {
		urlHealth := &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(node.Endpoint.ServiceHost, strconv.Itoa(int(node.Endpoint.ServicePort))),
			Path:   healthPath,
		}
		res, err := http.Get(urlHealth.String())
		if err != nil {
			return err
		} else {
			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("Bad status from %s:%d", node.Endpoint.ServiceHost, node.Endpoint.ServicePort)
			}
			return nil
		}
	} else {
		return nil
	}
}
