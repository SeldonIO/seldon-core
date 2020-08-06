package predictor

import (
	"fmt"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"net"
)

func Ready(node *v1.PredictiveUnit) error {
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
