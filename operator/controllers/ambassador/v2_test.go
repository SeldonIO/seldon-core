package ambassador

import (
	"fmt"
	v2 "github.com/emissary-ingress/emissary/v3/pkg/api/getambassador.io/v2"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestGetV2Mapping(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                string
		isREST              bool
		mlDep               *machinelearningv1.SeldonDeployment
		p                   *machinelearningv1.PredictorSpec
		addNamespace        bool
		serviceName         string
		serviceNameExternal string
		customHeader        string
		customRegexHeader   string
		weight              *int32
		shadowing           bool
		engine_port         int
		isExplainer         bool
		instance_id         string
		expected            *v2.Mapping
	}
	getInt32Ptr := func(val int32) *int32 { return &val }
	getIntPtr := func(val int) *int { return &val }
	getDuration := func(milli int) *v2.MillisecondDuration {
		return &v2.MillisecondDuration{
			Duration: time.Millisecond * time.Duration(milli),
		}
	}
	rewrite := "/"
	tests := []test{
		{
			name:   "basic",
			isREST: true,
			mlDep: &machinelearningv1.SeldonDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
			p: &machinelearningv1.PredictorSpec{
				Name: "predictor",
			},
			addNamespace:        false,
			serviceName:         "seldon",
			serviceNameExternal: "seldon",
			customHeader:        "",
			customRegexHeader:   "",
			weight:              getInt32Ptr(100),
			shadowing:           false,
			engine_port:         2000,
			isExplainer:         false,
			instance_id:         "1",
			expected: &v2.Mapping{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "seldon_test_predictor_rest_mapping",
					Namespace: "default",
				},
				Spec: v2.MappingSpec{
					ClusterTag:   "seldon_http",
					AmbassadorID: v2.AmbassadorID{"1"},
					Prefix:       "/seldon/seldon/",
					Service:      "seldon.default:2000",
					Rewrite:      &rewrite,
					Timeout:      getDuration(3000),
					Weight:       getIntPtr(100),
					Headers:      map[string]v2.BoolOrString{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m, err := getV2Mapping(test.isREST,
				test.mlDep,
				test.p,
				test.addNamespace,
				test.serviceName,
				test.serviceNameExternal,
				test.customHeader,
				test.customRegexHeader,
				test.weight,
				test.shadowing,
				test.engine_port,
				test.isExplainer,
				test.instance_id)
			g.Expect(err).To(BeNil())
			g.Expect(m.Spec).To(Equal(test.expected.Spec))
			fmt.Printf("%v", *m)
		})

	}

}
