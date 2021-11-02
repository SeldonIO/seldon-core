package predictor

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

func TestGetPredictorFromEnv(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	var b64Error base64.CorruptInputError

	tests := []struct {
		name      string
		key       string
		val       string
		predictor *v1.PredictorSpec
		err       error
	}{
		{
			name: "missing env var",
			err:  fmt.Errorf("Predictor not found, enviroment variable %s not set", EnvKeyEnginePredictor),
		},
		{
			name: "empty value",
			key:  EnvKeyEnginePredictor,
		},
		{
			name: "non-b64 value",
			key:  EnvKeyEnginePredictor,
			val:  ":;,",
			err:  b64Error,
		},
	}

	// unset existing env var & reset at end of test
	val, ok := os.LookupEnv(EnvKeyEnginePredictor)
	if ok {
		if err := os.Unsetenv(EnvKeyEnginePredictor); err != nil {
			t.Fatalf("failed to unset env var %v: %v", EnvKeyEnginePredictor, err)
		}
		defer func() {
			os.Setenv(EnvKeyEnginePredictor, val)
		}()
	}

	setenv := func(key, val string) {
		if key == "" {
			return
		}
		if err := os.Setenv(key, val); err != nil {
			t.Fatalf("failed to set env var: %v", err)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setenv(tt.key, tt.val)
			got, err := getPredictorFromEnv()
			g.Expect(got).Should(Equal(tt.predictor))
			if tt.err == nil {
				g.Expect(err).Should(BeNil())
			} else {
				g.Expect(err).Should(Equal(tt.err))
			}
		})
	}
}
