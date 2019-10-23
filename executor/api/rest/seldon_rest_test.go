package rest

import (
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/api/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAliveEndpoint(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

	r := NewSeldonRestApi(nil, nil, true)
	r.Initialise()

	req, _ := http.NewRequest("GET", "/live", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)

	g.Expect(res.Code).To(gomega.Equal(200))
}

func TestSimpleModel(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

	model := v1alpha2.MODEL
	p := v1alpha2.PredictorSpec{
		Name: "p",
		Graph: &v1alpha2.PredictiveUnit{
			Type: &model,
			Endpoint: &v1alpha2.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1alpha2.REST,
			},
		},
	}

	r := NewSeldonRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))
}
