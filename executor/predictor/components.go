package predictor

import (
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"math/rand"
	"strconv"
)

func (p *PredictorProcess) abTestRouter(node *v1alpha2.PredictiveUnit) (int, error) {
	ratioA, err := strconv.ParseFloat(node.Parameters[0].Value, 32)
	if err != nil {
		return 0, err
	}
	if rand.Float64() < ratioA {
		return 0, nil
	} else {
		return 1, nil
	}
}
