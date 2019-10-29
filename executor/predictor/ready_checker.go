package predictor

import (
	"fmt"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"net"
)

func Ready(node *v1alpha2.PredictiveUnit) error {
	for _, child := range node.Children {
		err := Ready(&child)
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
