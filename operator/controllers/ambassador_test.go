package controllers

import (
	"strings"
	"testing"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAmbassadorBasic(t *testing.T) {
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"}}
	p := machinelearningv1.PredictorSpec{Name: "p"}
	s, err := getAmbassadorConfigs(&mlDep, &p, "myservice", 9000, 5000, "")
	if err != nil {
		t.Fatalf("Config format error")
	}
	t.Logf("%s\n\n", s)
	parts := strings.Split(s, "---\n")[1:]

	if len(parts) != 2 {
		t.Fatalf("Bad number of configs returned %d", len(parts))
	}

	for _, part := range parts {
		c := AmbassadorConfig{}
		t.Logf("Config: %s", part)

		err = yaml.Unmarshal([]byte(s), &c)
		if err != nil {
			t.Fatalf("Failed to unmarshall")
		}

		if len(c.Headers) > 0 {
			t.Fatalf("Found headers")
		}
		if c.Prefix != "/seldon/default/mymodel/" {
			t.Fatalf("Found bad prefix %s", c.Prefix)
		}

		if c.InstanceId != "" {
			t.Fatalf("Found ambassador_id %s", c.InstanceId)
		}
	}

}

func TestAmbassadorID(t *testing.T) {
	mlDep := machinelearningv1.SeldonDeployment{ObjectMeta: metav1.ObjectMeta{Name: "mymodel"},
		Spec: machinelearningv1.SeldonDeploymentSpec{Annotations: map[string]string{ANNOTATION_AMBASSADOR_ID: "myinstance_id"}}}
	p := machinelearningv1.PredictorSpec{Name: "p"}
	s, err := getAmbassadorConfigs(&mlDep, &p, "myservice", 9000, 5000, "")
	if err != nil {
		t.Fatalf("Config format error")
	}
	t.Logf("%s\n\n", s)
	parts := strings.Split(s, "---\n")[1:]

	if len(parts) != 2 {
		t.Fatalf("Bad number of configs returned %d", len(parts))
	}

	for _, part := range parts {
		c := AmbassadorConfig{}
		t.Logf("Config: %s", part)

		err = yaml.Unmarshal([]byte(s), &c)
		if err != nil {
			t.Fatalf("Failed to unmarshall")
		}

		if len(c.Headers) > 0 {
			t.Fatalf("Found headers")
		}
		if c.Prefix != "/seldon/default/mymodel/" {
			t.Fatalf("Found bad prefix %s", c.Prefix)
		}

		if c.InstanceId != "myinstance_id" {
			t.Fatalf("Found mismatch ambassador_id %s", c.InstanceId)
		}
	}
}
