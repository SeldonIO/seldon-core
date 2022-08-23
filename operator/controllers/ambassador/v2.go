package ambassador

import (
	"fmt"
	v2 "github.com/emissary-ingress/emissary/v3/pkg/api/getambassador.io/v2"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	utils2 "github.com/seldonio/seldon-core/operator/controllers/utils"
	"github.com/seldonio/seldon-core/operator/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"time"
)

func getV2Mapping(isREST bool,
	mlDep *machinelearningv1.SeldonDeployment,
	p *machinelearningv1.PredictorSpec,
	addNamespace bool,
	serviceName string,
	serviceNameExternal string,
	customHeader string,
	customRegexHeader string,
	weight *int32,
	shadowing bool,
	engine_port int,
	isExplainer bool,
	instance_id string) (*v2.Mapping, error) {

	namespace := utils2.GetNamespace(mlDep)
	seldonId := machinelearningv1.GetSeldonDeploymentName(mlDep)

	// Set timeout
	timeout, err := strconv.Atoi(utils2.GetAnnotation(mlDep, ANNOTATION_REST_TIMEOUT, "3000"))
	if err != nil {
		return nil, err
	}
	timeoutDur := &v2.MillisecondDuration{
		Duration: time.Millisecond * time.Duration(timeout),
	}

	var m *v2.Mapping
	coreName := p.Name
	if isREST {
		if isExplainer {
			coreName = p.Name + constants.ExplainerNameSuffix
			serviceNameExternal = serviceNameExternal + constants.ExplainerPathSuffix + "/" + p.Name
		}
		rewrite := "/"

		name := "seldon_" + mlDep.ObjectMeta.Name + "_" + coreName + "_rest_mapping"
		prefix := "/seldon/" + serviceNameExternal + "/"
		if addNamespace {
			name = "seldon-" + namespace + "-" + mlDep.ObjectMeta.Name + "-" + coreName + "-rest"
			prefix = "/seldon/" + namespace + "/" + serviceNameExternal + "/"
		}

		m = &v2.Mapping{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{machinelearningv1.Label_seldon_id: seldonId},
			},
			Spec: v2.MappingSpec{
				Prefix:  prefix,
				Rewrite: &rewrite,
				Headers: map[string]v2.BoolOrString{},
				Service: serviceName + "." + namespace + ":" + strconv.Itoa(engine_port),
				Timeout: timeoutDur,
			},
		}

	} else {
		if isExplainer {
			coreName = p.Name + constants.ExplainerNameSuffix
			serviceNameExternal = serviceNameExternal + constants.ExplainerPathSuffix
		}
		name := "seldon-" + mlDep.ObjectMeta.Name + "-" + coreName + "-grpc"
		rewrite := ""
		grpc := true

		m = &v2.Mapping{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{machinelearningv1.Label_seldon_id: seldonId},
			},
			Spec: v2.MappingSpec{
				Prefix:      constants.GRPCRegExMatchAmbassador,
				GRPC:        &grpc,
				PrefixRegex: &grpc,
				Rewrite:     &rewrite,
				Headers:     map[string]v2.BoolOrString{"seldon": {String: &serviceNameExternal}},
				Service:     serviceName + "." + namespace + ":" + strconv.Itoa(engine_port),
				Timeout:     timeoutDur,
			},
		}
	}

	retries, err := strconv.Atoi(utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_RETRIES, AMBASSADOR_DEFAULT_RETRIES))
	if err != nil {
		return nil, err
	}
	if retries != 0 {
		m.Spec.RetryPolicy = &v2.RetryPolicy{
			RetryOn:    "gateway-error",
			NumRetries: &retries,
		}
	}

	if weight != nil {
		weightInt := int(*weight)
		m.Spec.Weight = &weightInt
	}

	if timeout > AMBASSADOR_IDLE_TIMEOUT {
		timeoutIdle := &v2.MillisecondDuration{
			Duration: time.Millisecond * time.Duration(timeout),
		}
		m.Spec.IdleTimeout = timeoutIdle
	}

	if customHeader != "" {
		parts := strings.Split(customHeader, ":")
		var val v2.BoolOrString
		switch len(parts) {
		case 1:
			exists := true
			val = v2.BoolOrString{
				Bool: &exists,
			}
		case 2:
			trimmed := strings.TrimSpace(parts[1])
			val = v2.BoolOrString{
				String: &trimmed,
			}
		default:
			return nil, fmt.Errorf("Only a single custom header match is allowed at present but you provided: %s", customHeader)
		}
		key := strings.TrimSpace(parts[0])
		m.Spec.Headers[key] = val
	}

	if customRegexHeader != "" {
		headers := strings.Split(customHeader, ":")
		elementMap := make(map[string]string)
		for i := 0; i < len(headers); i += 2 {
			key := strings.TrimSpace(headers[i])
			val := strings.TrimSpace(headers[i+1])
			elementMap[key] = val
		}
		m.Spec.RegexHeaders = elementMap
	}

	if shadowing {
		m.Spec.Shadow = &shadowing
	}

	if instance_id != "" {
		m.Spec.AmbassadorID = []string{instance_id}
	}

	circuitBreakerConfig, err := getV2AmbassadorCircuitBreakerConfig(mlDep)
	if err != nil {
		return nil, err
	}
	if circuitBreakerConfig != nil {
		m.Spec.CircuitBreakers = []*v2.CircuitBreaker{
			circuitBreakerConfig,
		}
	}

	if !utils.IsEmptyTLS(p) {
		tlsConfigName := getTLSConfigName(mlDep, p, addNamespace)
		m.Spec.TLS = &v2.BoolOrString{
			String: &tlsConfigName,
		}
	}

	return m, nil
}

func getV2AmbassadorCircuitBreakerConfig(
	mlDep *machinelearningv1.SeldonDeployment,
) (*v2.CircuitBreaker, error) {

	circuitBreakersMaxConnections := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_CONNECTIONS, "")
	circuitBreakersMaxPendingRequests := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_PENDING_REQUESTS, "")
	circuitBreakersMaxRequests := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_REQUESTS, "")
	circuitBreakersMaxRetries := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_RETRIES, "")

	// circuit breaker exists
	if circuitBreakersMaxConnections != "" ||
		circuitBreakersMaxPendingRequests != "" ||
		circuitBreakersMaxRequests != "" ||
		circuitBreakersMaxRetries != "" {

		circuitBreaker := v2.CircuitBreaker{}

		if circuitBreakersMaxConnections != "" {
			maxConnections, err := strconv.Atoi(circuitBreakersMaxConnections)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxConnections = &maxConnections
		}
		if circuitBreakersMaxPendingRequests != "" {
			maxPendingRequests, err := strconv.Atoi(circuitBreakersMaxPendingRequests)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxPendingRequests = &maxPendingRequests
		}
		if circuitBreakersMaxRequests != "" {
			maxRequests, err := strconv.Atoi(circuitBreakersMaxRequests)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxRequests = &maxRequests
		}
		if circuitBreakersMaxRetries != "" {
			maxRetries, err := strconv.Atoi(circuitBreakersMaxRetries)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxRetries = &maxRetries
		}
		return &circuitBreaker, nil
	}
	return nil, nil
}

func getTLSConfigName(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, addNamespace bool) string {
	name := p.Name
	name = "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config"
	if addNamespace {
		namespace := utils2.GetNamespace(mlDep)
		name = "seldon_" + namespace + "_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config"
	}
	return name
}

func getV2TLSConfig(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, addNamespace bool) *v2.TLSContext {
	seldonId := machinelearningv1.GetSeldonDeploymentName(mlDep)
	namespace := utils2.GetNamespace(mlDep)

	tlsConfig := v2.TLSContext{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getTLSConfigName(mlDep, p, addNamespace),
			Namespace: namespace,
			Labels:    map[string]string{machinelearningv1.Label_seldon_id: seldonId},
		},
		Spec: v2.TLSContextSpec{
			Hosts:  []string{},
			Secret: p.SSL.CertSecretName,
		},
	}
	return &tlsConfig
}

func GetV2AmbassadorConfigs(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, serviceName string, engine_http_port, engine_grpc_port int, isExplainer bool) ([]*v2.Mapping, []*v2.TLSContext, error) {
	var weight *int32
	// Ignore weight on first predictor and let Ambassador handle this
	if p.Shadow || mlDep.Spec.Predictors[0].Name != p.Name {
		weight = &p.Traffic
	}

	shadowing := p.Shadow
	serviceNameExternal := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_SERVICE, mlDep.GetName())
	customHeader := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_HEADER, "")
	customRegexHeader := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_REGEX_HEADER, "")
	instance_id := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_ID, "")

	cRestGlobal, err := getV2Mapping(true, mlDep, p, true, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_http_port, isExplainer, instance_id)
	if err != nil {
		return nil, nil, err
	}

	cGrpcGlobal, err := getV2Mapping(false, mlDep, p, true, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_grpc_port, isExplainer, instance_id)
	if err != nil {
		return nil, nil, err
	}

	cRestNamespaced, err := getV2Mapping(true, mlDep, p, false, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_http_port, isExplainer, instance_id)
	if err != nil {
		return nil, nil, err
	}

	cGrpcNamespaced, err := getV2Mapping(false, mlDep, p, false, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_grpc_port, isExplainer, instance_id)
	if err != nil {
		return nil, nil, err
	}

	if utils.IsEmptyTLS(p) {
		if utils.GetEnv("AMBASSADOR_SINGLE_NAMESPACE", "false") == "true" {
			return []*v2.Mapping{cRestGlobal, cGrpcGlobal, cRestNamespaced, cGrpcNamespaced}, []*v2.TLSContext{}, nil
		} else {
			return []*v2.Mapping{cRestGlobal, cGrpcGlobal}, []*v2.TLSContext{}, nil
		}
	} else {
		cTLSGlobal := getV2TLSConfig(mlDep, p, false)
		cTLSNamespaced := getV2TLSConfig(mlDep, p, true)

		if utils.GetEnv("AMBASSADOR_SINGLE_NAMESPACE", "false") == "true" {
			return []*v2.Mapping{cRestGlobal, cGrpcGlobal, cRestNamespaced, cGrpcNamespaced}, []*v2.TLSContext{cTLSGlobal, cTLSNamespaced}, nil
		} else {
			return []*v2.Mapping{cRestGlobal, cGrpcGlobal}, []*v2.TLSContext{cTLSGlobal}, nil
		}
	}
}
