package v1alpha1

import (
	. "github.com/onsi/gomega"
	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler/api"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestAsModelDetails(t *testing.T) {
	t.Logf("Started")
	logrus.SetLevel(logrus.DebugLevel)
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		model        *Model
		modelDetails *scheduler.ModelDetails
		error        bool
	}
	replicas := int32(4)
	secret := "secret"
	modelType := "sklearn"
	server := "server"
	m1 := resource.MustParse("1M")
	m1bytes := uint64(100000)
	tests := []test{
		{
			name: "simple",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					ResourceVersion: "1",
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
				},
			},
			modelDetails: &scheduler.ModelDetails{
				Name:     "foo",
				Version:  "1",
				Uri:      "gs://test",
				Replicas: 1,
			},
		},
		{
			name: "complex",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					ResourceVersion: "1",
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						ModelType:  &modelType,
						StorageURI: "gs://test",
						SecretName: &secret,
					},
					Logger:       &LoggingSpec{},
					Requirements: []string{"a", "b"},
					Replicas:     &replicas,
					Server:       &server,
				},
			},
			modelDetails: &scheduler.ModelDetails{
				Name:          "foo",
				Version:       "1",
				Uri:           "gs://test",
				Replicas:      4,
				Requirements:  []string{"a", "b", modelType},
				StorageConfig: &scheduler.StorageConfig{Config: &scheduler.StorageConfig_StorageSecretName{StorageSecretName: "secret"}},
				Server:        &server,
				LogPayloads:   true,
			},
		},
		{
			name: "memory",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					ResourceVersion: "1",
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
					Memory: &m1,
				},
			},
			modelDetails: &scheduler.ModelDetails{
				Name:        "foo",
				Version:     "1",
				Uri:         "gs://test",
				Replicas:    1,
				MemoryBytes: &m1bytes,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			md, err := test.model.AsModelDetails()
			if !test.error {
				g.Expect(err).To(BeNil())
				g.Expect(md).To(Equal(test.modelDetails))
			} else {
				g.Expect(err).ToNot(BeNil())
			}

		})
	}
}
