package ambassador

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TEST_DEFAULT_EXPECTED_RETRIES = 0
)

func basicAmbassadorTests(t *testing.T, mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, expectedWeight int32, expectedInstanceId string, expectedRetries int, isExplainer bool, isTLS bool) {
	g := NewGomegaWithT(t)
	s, err := GetAmbassadorConfigs(mlDep, p, "myservice", 9000, 5000, isExplainer)
	g.Expect(err).To(BeNil())
	parts := strings.Split(s, "---\n")[1:]
	g.Expect(len(parts)).To(Equal(3))
	c := AmbassadorConfig{}
	err = yaml.Unmarshal([]byte(parts[0]), &c)
	g.Expect(err).To(BeNil())
	if isExplainer {
		g.Expect(c.Prefix).To(Equal("/seldon/default/mymodel" + constants.ExplainerPathSuffix + "/" + p.Name + "/"))
	} else {
		g.Expect(c.Prefix).To(Equal("/seldon/default/mymodel/"))
	}

	g.Expect(c.Weight).To(Equal(expectedWeight))
	g.Expect(c.InstanceId).To(Equal(expectedInstanceId))
	if expectedRetries > 0 {
		g.Expect(c.RetryPolicy.NumRetries).To(Equal(expectedRetries))
	} else {
		g.Expect(c.RetryPolicy).To(BeNil())
	}

	if isTLS {
		g.Expect(len(c.TLS)).ToNot(Equal(0))
	} else {
		g.Expect(len(c.TLS)).To(Equal(0))
	}
}

func TestAmbassadorTLS(t *testing.T) {
	p1 := machinelearningv1.PredictorSpec{
		Name: "p1",
		SSL: &machinelearningv1.SSL{
			CertSecretName: "model-secret-name",
		},
	}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				p1,
			},
		},
	}

	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, false, true)
}

func TestAmbassadorSingle(t *testing.T) {
	p1 := machinelearningv1.PredictorSpec{Name: "p1"}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				p1,
			},
		},
	}

	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, true, false)
}

func TestAmbassadorCanary(t *testing.T) {
	p1 := machinelearningv1.PredictorSpec{Name: "p1", Traffic: 20}
	p2 := machinelearningv1.PredictorSpec{Name: "p2", Traffic: 80}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				p1,
				p2,
			},
		},
	}

	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p2, 80, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, true, false)
	basicAmbassadorTests(t, &mlDep, &p2, 80, "", TEST_DEFAULT_EXPECTED_RETRIES, true, false)
}

func TestAmbassadorCanaryEqual(t *testing.T) {
	p1 := machinelearningv1.PredictorSpec{Name: "p1", Traffic: 50}
	p2 := machinelearningv1.PredictorSpec{Name: "p2", Traffic: 50}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				p1,
				p2,
			},
		},
	}

	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p2, 50, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
}

func TestAmbassadorCanaryThree(t *testing.T) {
	p1 := machinelearningv1.PredictorSpec{Name: "p1", Traffic: 60}
	p2 := machinelearningv1.PredictorSpec{Name: "p2", Traffic: 20}
	p3 := machinelearningv1.PredictorSpec{Name: "p3", Traffic: 20}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				p1,
				p2,
				p3,
			},
		},
	}

	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p2, 20, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p3, 20, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
}

func TestAmbassadorCanaryThreeEqual(t *testing.T) {
	p1 := machinelearningv1.PredictorSpec{Name: "p1", Traffic: 33}
	p2 := machinelearningv1.PredictorSpec{Name: "p2", Traffic: 33}
	p3 := machinelearningv1.PredictorSpec{Name: "p3", Traffic: 33}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				p1,
				p2,
				p3,
			},
		},
	}

	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p2, 33, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
	basicAmbassadorTests(t, &mlDep, &p3, 33, "", TEST_DEFAULT_EXPECTED_RETRIES, false, false)
}

func TestAmbassadorID(t *testing.T) {
	const instanceId = "myinstance_id"
	p1 := machinelearningv1.PredictorSpec{Name: "p"}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Annotations: map[string]string{ANNOTATION_AMBASSADOR_ID: instanceId},
			Predictors: []machinelearningv1.PredictorSpec{
				p1,
			},
		},
	}
	basicAmbassadorTests(t, &mlDep, &p1, 0, instanceId, TEST_DEFAULT_EXPECTED_RETRIES, false, false)
}

func TestAmbassadorRetriesAnnotation(t *testing.T) {
	p := machinelearningv1.PredictorSpec{Name: "p"}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Annotations: map[string]string{ANNOTATION_AMBASSADOR_RETRIES: "2"},
			Predictors: []machinelearningv1.PredictorSpec{
				p,
			},
		},
	}
	basicAmbassadorTests(t, &mlDep, &p, 0, "", 2, false, false)
}

func circuitBreakerAmbassadorTests(t *testing.T,
	mlDep *machinelearningv1.SeldonDeployment,
	p *machinelearningv1.PredictorSpec,
	expectedNumCircuitBreaker int,
	expectedMaxConnections int,
	expectedMaxPendingRequests int,
	expectedMaxRequests int,
	expectedMaxRetries int,
) {
	g := NewGomegaWithT(t)
	s, err := GetAmbassadorConfigs(mlDep, p, "myservice", 9000, 5000, false)
	g.Expect(err).To(BeNil())
	parts := strings.Split(s, "---\n")[1:]
	g.Expect(len(parts)).To(Equal(3))
	c := AmbassadorConfig{}
	err = yaml.Unmarshal([]byte(parts[0]), &c)
	g.Expect(err).To(BeNil())
	g.Expect(c.Prefix).To(Equal("/seldon/default/mymodel/"))

	g.Expect(len(c.CircuitBreakers)).To(Equal(expectedNumCircuitBreaker))
	if expectedNumCircuitBreaker > 0 {
		g.Expect(c.CircuitBreakers[0].MaxConnections).To(Equal(expectedMaxConnections))
		g.Expect(c.CircuitBreakers[0].MaxPendingRequests).To(Equal(expectedMaxPendingRequests))
		g.Expect(c.CircuitBreakers[0].MaxRequests).To(Equal(expectedMaxRequests))
		g.Expect(c.CircuitBreakers[0].MaxRetries).To(Equal(expectedMaxRetries))
	}
}

func TestAmbassadorNoCircuitBreakerAnnotation(t *testing.T) {
	p := machinelearningv1.PredictorSpec{Name: "p"}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				p,
			},
		},
	}
	circuitBreakerAmbassadorTests(t, &mlDep, &p, 0, 0, 0, 0, 0)
}

func TestAmbassadorCircuitBreakerAnnotation(t *testing.T) {
	p := machinelearningv1.PredictorSpec{Name: "p"}
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Annotations: map[string]string{
				ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_CONNECTIONS:      "10",
				ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_PENDING_REQUESTS: "15",
				ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_REQUESTS:         "20",
				ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_RETRIES:          "5",
			},
			Predictors: []machinelearningv1.PredictorSpec{
				p,
			},
		},
	}
	circuitBreakerAmbassadorTests(t, &mlDep, &p, 1, 10, 15, 20, 5)
}
