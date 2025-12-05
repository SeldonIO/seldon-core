package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/godog/k8sclient"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	v "github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/watch"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type World struct {
	namespace            string
	kubeClient           k8sclient.Client
	StartingClusterState string //todo: this will be a combination of starting state awareness of core 2 such as the
	//todo:  server config,seldon config and seldon runtime to be able to reconcile to starting state should we change
	//todo: the state such as reducing replicas to 0 of scheduler to test unavailability
	CurrentModel       *Model
	Models             map[string]*Model
	k8sClient          v.Interface
	httpClient         *http.Client
	gatewayHTTPAddress string
	lastHTTPResponse   *http.Response
}

func NewWorld(namespace string) *World {
	k8sClient, err := v.NewForConfig(controllerruntime.GetConfigOrDie())
	if err != nil {
		panic(err)
	}

	return &World{
		namespace:          namespace,
		k8sClient:          k8sClient,
		httpClient:         &http.Client{},
		gatewayHTTPAddress: "http://localhost:9000/",
	}
}

func (w *World) sendHTTPInferenceRequest(timeout, model string, payload *godog.DocString) error {
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout %s: %w", timeout, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%sv2/models/%s/infer", w.gatewayHTTPAddress, model), strings.NewReader(payload.Content))
	if err != nil {
		return fmt.Errorf("could not create http request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", "seldon-mesh.inference.seldon")
	req.Header.Add("Seldon-model", model)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send http request: %w", err)
	}
	w.lastHTTPResponse = resp
	return nil
}

func isSubset(needle, hay any) bool {
	nObj, nOK := needle.(map[string]any)
	hObj, hOK := hay.(map[string]any)
	if nOK && hOK {
		for k, nv := range nObj {
			hv, exists := hObj[k]
			if !exists || !isSubset(nv, hv) {
				return false
			}
		}
		return true
	}

	return reflect.DeepEqual(needle, hay)
}

func containsSubset(needle, hay any) bool {
	if isSubset(needle, hay) {
		return true
	}
	switch h := hay.(type) {
	case map[string]any:
		for _, v := range h {
			if containsSubset(needle, v) {
				return true
			}
		}
	case []any:
		for _, v := range h {
			if containsSubset(needle, v) {
				return true
			}
		}
	}
	return false
}

func jsonContainsObjectSubset(jsonStr, needleStr string) (bool, error) {
	var hay, needle any
	if err := json.Unmarshal([]byte(jsonStr), &hay); err != nil {
		return false, fmt.Errorf("could not unmarshal hay json %s: %w", jsonStr, err)
	}
	if err := json.Unmarshal([]byte(needleStr), &needle); err != nil {
		return false, fmt.Errorf("could not unmarshal needle json %s: %w", needleStr, err)
	}
	return containsSubset(needle, hay), nil
}

func (w *World) httpRespCheckBodyContainsJSON(expectJSON *godog.DocString) error {
	if w.lastHTTPResponse == nil {
		return errors.New("no http response found")
	}

	body, err := io.ReadAll(w.lastHTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	ok, err := jsonContainsObjectSubset(string(body), expectJSON.Content)
	if err != nil {
		return fmt.Errorf("could not check if json contains object: %w", err)
	}

	if !ok {
		return fmt.Errorf("%s does not contain %s", string(body), expectJSON)
	}

	return nil
}
func (w *World) httpRespCheckStatus(status int) error {
	if w.lastHTTPResponse == nil {
		return errors.New("no http response found")
	}
	if status != w.lastHTTPResponse.StatusCode {
		return fmt.Errorf("expected http response status code %d, got %d", status, w.lastHTTPResponse.StatusCode)
	}
	return nil
}

func (w *World) deployModelSpec(spec *godog.DocString) error {
	modelSpec := &v1alpha1.Model{}
	if err := yaml.Unmarshal([]byte(spec.Content), &modelSpec); err != nil {
		return fmt.Errorf("failed unmarshalling model spec: %w", err)
	}

	if _, err := w.k8sClient.MlopsV1alpha1().Models(w.namespace).Create(context.TODO(), modelSpec, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("failed creating model: %w", err)
	}
	return nil
}

func (w *World) waitForModelReady(model, timeout string) error {
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout %s: %w", timeout, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	foundModel, err := w.k8sClient.MlopsV1alpha1().Models(w.namespace).Get(ctx, model, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed getting model: %w", err)
	}

	if foundModel.Status.IsReady() {
		return nil
	}

	watcher, err := w.k8sClient.MlopsV1alpha1().Models(w.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", model),
		ResourceVersion: foundModel.ResourceVersion,
		Watch:           true,
	})
	if err != nil {
		return fmt.Errorf("failed subscribed to watch model: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}

			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}

			if event.Type == watch.Added || event.Type == watch.Modified {
				model := event.Object.(*v1alpha1.Model)
				if model.Status.IsReady() {
					return nil
				}
			}

			if event.Type == watch.Deleted {
				return fmt.Errorf("resource was deleted")
			}
		}
	}
}
