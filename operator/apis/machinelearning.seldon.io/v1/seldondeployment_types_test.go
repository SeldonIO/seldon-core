package v1

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDeploymentNameOneModel(t *testing.T) {
	g := NewGomegaWithT(t)
	mldepName := "dep"
	predictorName := "p1"
	modelName := "classifier"
	mlDep := &SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mldepName,
			Namespace: "default",
		},
		Spec: SeldonDeploymentSpec{
			Protocol: "abc",
			Predictors: []PredictorSpec{
				{
					Name: predictorName,
					ComponentSpecs: []*SeldonPodSpec{
						{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "seldonio/mock_classifier:1.0",
										Name:  modelName,
									},
								},
							},
						},
					},
					Graph: PredictiveUnit{
						Name: modelName,
					},
				},
			},
		},
	}

	name := GetDeploymentName(mlDep, mlDep.Spec.Predictors[0], mlDep.Spec.Predictors[0].ComponentSpecs[0], 0)
	parts := strings.Split(name, "-")
	g.Expect(len(parts)).To(Equal(4))
	g.Expect(parts[0]).To(Equal(mldepName))
	g.Expect(parts[1]).To(Equal(predictorName))
	g.Expect(parts[2]).To(Equal("0"))
	g.Expect(parts[3]).To(Equal(modelName))
}

func TestGetDeploymentNameLong(t *testing.T) {
	g := NewGomegaWithT(t)
	mldepName := "dep"
	predictorName := "p1"
	modelName := "C12345678912345678901234567890123456789012345678901234567890"
	mlDep := &SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mldepName,
			Namespace: "default",
		},
		Spec: SeldonDeploymentSpec{
			Protocol: "abc",
			Predictors: []PredictorSpec{
				{
					Name: predictorName,
					ComponentSpecs: []*SeldonPodSpec{
						{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "seldonio/mock_classifier:1.0",
										Name:  modelName,
									},
								},
							},
						},
					},
					Graph: PredictiveUnit{
						Name: modelName,
					},
				},
			},
		},
	}

	name := GetDeploymentName(mlDep, mlDep.Spec.Predictors[0], mlDep.Spec.Predictors[0].ComponentSpecs[0], 0)
	parts := strings.Split(name, "-")
	g.Expect(len(parts)).To(Equal(2))
	g.Expect(parts[0]).To(Equal(DeploymentNamePrefix))
}
