package agent

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

const (
	backEndServerPort = 8088
)

func v2_infer(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	model_name := params["model_name"]
	_, _ = w.Write([]byte("Model inference: " + model_name))
}

func v2_load(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	model_name := params["model_name"]
	_, _ = w.Write([]byte("Model load: " + model_name))
}

func v2_unload(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	model_name := params["model_name"]
	_, _ = w.Write([]byte("Model unload: " + model_name))
}

func setupMockMLServer() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/v2/models/{model_name:\\w+}/infer", v2_infer).Methods("POST")
	rtr.HandleFunc("/v2/repository/models/{model_name:\\w+}/load", v2_load).Methods("POST")
	rtr.HandleFunc("/v2/repository/models/{model_name:\\w+}/unload", v2_unload).Methods("POST")

	http.Handle("/", rtr)

	if err := http.ListenAndServe(":"+strconv.Itoa(backEndServerPort), nil); err != nil {
		log.Fatal(err)
	}
}

func setupReverseProxy(logger log.FieldLogger, numModels int, modelPrefix string) *reverseHTTPProxy {
	v2Client := NewV2Client("localhost", backEndServerPort, logger)
	localCacheManager := setupLocalTestManager(numModels, modelPrefix, v2Client, numModels-2)
	rp := NewReverseHTTPProxy(logger, ReverseProxyHTTPPort)
	rp.SetState(localCacheManager)
	return rp
}

func TestReverseProxySmoke(t *testing.T) {

	g := NewGomegaWithT(t)
	dummyModelNamePrefix := "dummy_model"

	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	go setupMockMLServer()
	rpHTTP := setupReverseProxy(logger, 3, dummyModelNamePrefix)
	if err := rpHTTP.Start(); err != nil {
		t.Errorf("Cannot start reverse proxy %s", err)
	}

	t.Log("Testing model found")

	// load model
	rpHTTP.stateManager.modelVersions.addModelVersion(
		getDummyModelDetails(dummyModelNamePrefix+"_0", uint64(1), uint32(1)))

	// make a dummy predict call
	inferV2Path := "/v2/models/" + dummyModelNamePrefix + "_0" + "/infer"
	resp, err := http.Post(
		"http://localhost:"+strconv.Itoa(ReverseProxyHTTPPort)+inferV2Path,
		"application/json",
		nil)
	if err != nil || resp.StatusCode != 200 {
		t.Fatal("error")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyString := string(bodyBytes)

	if !strings.Contains(bodyString, dummyModelNamePrefix+"_0") {
		t.Fatal("Fail!!")
	}

	t.Logf("Testing model not found")
	// then make a call, this should fail
	// model_1 should not be loaded
	inferV2Path = "/v2/models/" + dummyModelNamePrefix + "_1" + "/infer"
	resp, err = http.Post(
		"http://localhost:"+strconv.Itoa(ReverseProxyHTTPPort)+inferV2Path,
		"application/json",
		nil)
	if err != nil {
		t.Fatal("error")
	}
	defer resp.Body.Close()
	g.Expect(resp.StatusCode).To(Equal(404))

	t.Log("Testing status")
	g.Expect(rpHTTP.Ready()).To(BeNil())
	_ = rpHTTP.Stop()
	g.Expect(rpHTTP.Ready()).NotTo(BeNil())

	t.Logf("Done!")
}

func TestExtractModelNamefromPath(t *testing.T) {
	t.Logf("Start!")

	g := NewGomegaWithT(t)

	type test struct {
		name     string
		path     string
		expected string
	}
	tests := []test{
		{
			name:     "noversion",
			path:     "v2/models/dummy_model/infer",
			expected: "dummy_model",
		},
		{
			name:     "withversion",
			path:     "v2/models/dummy_model/versions/1/infer",
			expected: "dummy_model",
		},
		{
			name:     "bad",
			path:     "dummy",
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model, err := ExtractModelNamefromPath(test.path)
			g.Expect(model).To(Equal(test.expected))
			if model == "" {
				g.Expect(err).NotTo(Equal(BeNil()))
			}
		})
	}

	t.Logf("Done!")
}
