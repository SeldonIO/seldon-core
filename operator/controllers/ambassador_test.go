package controllers

import (
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
)

const (
	TEST_DEFAULT_EXPECTED_RETRIES = 0
)

func basicAmbassadorTests(t *testing.T, mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, expectedWeight int32, expectedInstanceId string, expectedRetries int) {
	g := NewGomegaWithT(t)
	s, err := getAmbassadorConfigs(mlDep, p, "myservice", 9000, 5000, "")
	g.Expect(err).To(BeNil())
	parts := strings.Split(s, "---\n")[1:]
	g.Expect(len(parts)).To(Equal(2))
	c := AmbassadorConfig{}
	err = yaml.Unmarshal([]byte(parts[0]), &c)
	g.Expect(err).To(BeNil())
	g.Expect(c.Prefix).To(Equal("/seldon/default/mymodel/"))
	g.Expect(c.Weight).To(Equal(expectedWeight))
	g.Expect(c.InstanceId).To(Equal(expectedInstanceId))
	if expectedRetries > 0 {
		g.Expect(c.RetryPolicy.NumRetries).To(Equal(expectedRetries))
	} else {
		g.Expect(c.RetryPolicy).To(BeNil())
	}

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
	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES)
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
	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES)
	basicAmbassadorTests(t, &mlDep, &p2, 80, "", TEST_DEFAULT_EXPECTED_RETRIES)
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
	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES)
	basicAmbassadorTests(t, &mlDep, &p2, 50, "", TEST_DEFAULT_EXPECTED_RETRIES)
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
	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES)
	basicAmbassadorTests(t, &mlDep, &p2, 20, "", TEST_DEFAULT_EXPECTED_RETRIES)
	basicAmbassadorTests(t, &mlDep, &p3, 20, "", TEST_DEFAULT_EXPECTED_RETRIES)
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
	basicAmbassadorTests(t, &mlDep, &p1, 0, "", TEST_DEFAULT_EXPECTED_RETRIES)
	basicAmbassadorTests(t, &mlDep, &p2, 33, "", TEST_DEFAULT_EXPECTED_RETRIES)
	basicAmbassadorTests(t, &mlDep, &p3, 33, "", TEST_DEFAULT_EXPECTED_RETRIES)
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
	basicAmbassadorTests(t, &mlDep, &p1, 0, instanceId, TEST_DEFAULT_EXPECTED_RETRIES)
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
	basicAmbassadorTests(t, &mlDep, &p, 0, "", 2)
}
