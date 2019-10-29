package predictor

import (
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"math/rand"
	"strconv"
)

func (p *PredictorProcess) abTestRouter(node *v1alpha2.PredictiveUnit) (int, error) {
	ratioA := 0.5
	var err error
	if len(node.Parameters) == 1 && node.Parameters[0].Name == "ratioA" {
		ratioA, err = strconv.ParseFloat(node.Parameters[0].Value, 32)
		if err != nil {
			return 0, err
		}
	}

	if rand.Float64() < ratioA {
		return 0, nil
	} else {
		return 1, nil
	}
}
