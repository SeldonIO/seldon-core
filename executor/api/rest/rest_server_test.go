package rest

import (
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/test"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
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

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
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
