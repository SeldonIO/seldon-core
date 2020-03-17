package v1

import (
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestValidateBadProtocol(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Protocol: "abc",
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier",
								},
							},
						},
					},
				},
				Graph: &PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(2))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestValidateBadTransport(t *testing.T) {
	g := NewGomegaWithT(t)
	impl := MODEL
	spec := &SeldonDeploymentSpec{
		Transport: "abc",
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier",
								},
							},
						},
					},
				},
				Graph: &PredictiveUnit{
					Name: "classifier",
					Type: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(2))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestValidateMixedTransport(t *testing.T) {
	g := NewGomegaWithT(t)
	impl := MODEL
	spec := &SeldonDeploymentSpec{
		Transport: TransportRest,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier",
								},
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier2",
								},
							},
						},
					},
				},
				Graph: &PredictiveUnit{
					Name: "classifier",
					Type: &impl,
					Endpoint: &Endpoint{
						Type: GRPC,
					},
					Children: []PredictiveUnit{
						{
							Name: "classifier2",
							Type: &impl,
							Endpoint: &Endpoint{
								Type: REST,
							},
						},
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestValidateMixedMultipleTransport(t *testing.T) {
	g := NewGomegaWithT(t)
	impl := MODEL
	spec := &SeldonDeploymentSpec{
		Transport: TransportRest,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier",
								},
							},
						},
					},
				},
				Graph: &PredictiveUnit{
					Name: "classifier",
					Type: &impl,
					Endpoint: &Endpoint{
						Type: GRPC,
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).NotTo(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}
