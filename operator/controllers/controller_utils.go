package controllers

import (
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get the Namespace from the SeldonDeployment. Returns "default" if none found.
func getNamespace(deployment *machinelearningv1.SeldonDeployment) string {
	if len(deployment.ObjectMeta.Namespace) > 0 {
		return deployment.ObjectMeta.Namespace
	} else {
		return "default"
	}
}

// Translte the PredictorSpec p in to base64 encoded JSON to feed to engine in env var.
func getEngineVarJson(p *machinelearningv1.PredictorSpec) (string, error) {
	pcopy := p.DeepCopy()

	// engine doesn't need to know about metadata or explainer
	// leaving these out means they're not part of diffs on main predictor deployments
	for _, compSpec := range pcopy.ComponentSpecs {
		compSpec.Metadata.CreationTimestamp = &metav1.Time{}
	}
	pcopy.Explainer = nil

	str, err := json.Marshal(pcopy)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(str), nil
}

// Get an annotation from the Seldon Deployment given by annotationKey or return the fallback.
func getAnnotation(mlDep *machinelearningv1.SeldonDeployment, annotationKey string, fallback string) string {
	if annotation, hasAnnotation := mlDep.Spec.Annotations[annotationKey]; hasAnnotation {
		return annotation
	} else {
		if annotation, hasAnnotation := mlDep.Annotations[annotationKey]; hasAnnotation {
			return annotation
		} else {
			return fallback
		}
	}
}

//get annotations that start with seldon.io/engine
func getEngineEnvAnnotations(mlDep *machinelearningv1.SeldonDeployment) []corev1.EnvVar {

	envVars := make([]corev1.EnvVar, 0)
	var keys []string
	for k := range mlDep.Spec.Annotations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		//prefix indicates engine annotation but "seldon.io/engine-separate-pod" isn't an env one
		if strings.HasPrefix(k, "seldon.io/engine-") && k != machinelearningv1.ANNOTATION_SEPARATE_ENGINE {
			name := strings.TrimPrefix(k, "seldon.io/engine-")
			var replacer = strings.NewReplacer("-", "_")
			name = replacer.Replace(name)
			name = strings.ToUpper(name)
			envVars = append(envVars, corev1.EnvVar{Name: name, Value: mlDep.Spec.Annotations[k]})
		}
	}
	return envVars
}

// isEmptyExplainer will return true if the explainer can be considered empty
// (either by being nil or by having unset fields)
func isEmptyExplainer(explainer *machinelearningv1.Explainer) bool {
	return explainer == nil || explainer.Type == ""
}
