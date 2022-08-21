package ambassador

import (
	utils2 "github.com/seldonio/seldon-core/operator/controllers/utils"
	"strconv"
	"strings"

	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"gopkg.in/yaml.v2"
)

const (
	ANNOTATION_REST_TIMEOUT            = "seldon.io/rest-timeout"
	ANNOTATION_GRPC_TIMEOUT            = "seldon.io/grpc-timeout"
	ANNOTATION_AMBASSADOR_CUSTOM       = "seldon.io/ambassador-config"
	ANNOTATION_AMBASSADOR_SERVICE      = "seldon.io/ambassador-service-name"
	ANNOTATION_AMBASSADOR_HEADER       = "seldon.io/ambassador-header"
	ANNOTATION_AMBASSADOR_REGEX_HEADER = "seldon.io/ambassador-regex-header"
	ANNOTATION_AMBASSADOR_ID           = "seldon.io/ambassador-id"
	ANNOTATION_AMBASSADOR_RETRIES      = "seldon.io/ambassador-retries"

	ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_CONNECTIONS      = "seldon.io/ambassador-circuit-breakers-max-connections"
	ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_PENDING_REQUESTS = "seldon.io/ambassador-circuit-breakers-max-pending-requests"
	ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_REQUESTS         = "seldon.io/ambassador-circuit-breakers-max-requests"
	ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_RETRIES          = "seldon.io/ambassador-circuit-breakers-max-retries"

	YAML_SEP = "---\n"

	AMBASSADOR_IDLE_TIMEOUT    = 300000
	AMBASSADOR_DEFAULT_RETRIES = "0"
)

// AmbassadorCircuitBreakerConfig - struct for ambassador circuit breaker
type AmbassadorCircuitBreakerConfig struct {
	MaxConnections     int `yaml:"max_connections,omitempty"`
	MaxPendingRequests int `yaml:"max_pending_requests,omitempty"`
	MaxRequests        int `yaml:"max_requests,omitempty"`
	MaxRetries         int `yaml:"max_retries,omitempty"`
}

// Struct for Ambassador configuration
type AmbassadorTLSContextConfig struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Name       string   `yaml:"name"`
	Hosts      []string `yaml:"hosts"`
	Secret     string   `yaml:"secret"`
}

// Struct for Ambassador configuration
type AmbassadorConfig struct {
	ApiVersion      string                            `yaml:"apiVersion"`
	Kind            string                            `yaml:"kind"`
	Name            string                            `yaml:"name"`
	Grpc            *bool                             `yaml:"grpc,omitempty"`
	Prefix          string                            `yaml:"prefix"`
	PrefixRegex     *bool                             `yaml:"prefix_regex,omitempty"`
	Rewrite         string                            `yaml:"rewrite"`
	Service         string                            `yaml:"service"`
	TimeoutMs       int                               `yaml:"timeout_ms"`
	IdleTimeoutMs   *int                              `yaml:"idle_timeout_ms,omitempty"`
	Headers         map[string]string                 `yaml:"headers,omitempty"`
	RegexHeaders    map[string]string                 `yaml:"regex_headers,omitempty"`
	Weight          int32                             `yaml:"weight,omitempty"`
	Shadow          *bool                             `yaml:"shadow,omitempty"`
	RetryPolicy     *AmbassadorRetryPolicy            `yaml:"retry_policy,omitempty"`
	InstanceId      string                            `yaml:"ambassador_id,omitempty"`
	CircuitBreakers []*AmbassadorCircuitBreakerConfig `yaml:"circuit_breakers,omitempty"`
	TLS             string                            `yaml:"tls,omitempty"`
}

type AmbassadorRetryPolicy struct {
	RetryOn    string `yaml:"retry_on,omitempty"`
	NumRetries int    `yaml:"num_retries,omitempty"`
}

func getAmbassadorCircuitBreakerConfig(
	mlDep *machinelearningv1.SeldonDeployment,
) (*AmbassadorCircuitBreakerConfig, error) {

	circuitBreakersMaxConnections := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_CONNECTIONS, "")
	circuitBreakersMaxPendingRequests := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_PENDING_REQUESTS, "")
	circuitBreakersMaxRequests := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_REQUESTS, "")
	circuitBreakersMaxRetries := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CIRCUIT_BREAKING_MAX_RETRIES, "")

	// circuit breaker exists
	if circuitBreakersMaxConnections != "" ||
		circuitBreakersMaxPendingRequests != "" ||
		circuitBreakersMaxRequests != "" ||
		circuitBreakersMaxRetries != "" {

		circuitBreaker := AmbassadorCircuitBreakerConfig{}

		if circuitBreakersMaxConnections != "" {
			maxConnections, err := strconv.Atoi(circuitBreakersMaxConnections)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxConnections = maxConnections
		}
		if circuitBreakersMaxPendingRequests != "" {
			maxPendingRequests, err := strconv.Atoi(circuitBreakersMaxPendingRequests)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxPendingRequests = maxPendingRequests
		}
		if circuitBreakersMaxRequests != "" {
			maxRequests, err := strconv.Atoi(circuitBreakersMaxRequests)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxRequests = maxRequests
		}
		if circuitBreakersMaxRetries != "" {
			maxRetries, err := strconv.Atoi(circuitBreakersMaxRetries)
			if err != nil {
				return nil, err
			}
			circuitBreaker.MaxRetries = maxRetries
		}
		return &circuitBreaker, nil
	}
	return nil, nil
}

// Return TLSContext configuration if SSL is enabled with Ambassador and returns empty string if SSL is not enabled
func getAmbassadorTLSContextConfig(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, addNamespace bool) (string, error) {

	if utils.IsEmptyTLS(p) {
		return "", nil
	}

	name := p.Name
	namespace := utils2.GetNamespace(mlDep)

	c := AmbassadorTLSContextConfig{
		ApiVersion: "ambassador/v1",
		Kind:       "TLSContext",
		Name:       "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config",
		Hosts:      []string{},
		Secret:     p.SSL.CertSecretName,
	}

	if addNamespace {
		c.Name = "seldon_" + namespace + "_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config"
	}

	v, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// Return a REST configuration for Ambassador with optional custom settings.
func getAmbassadorRestConfig(mlDep *machinelearningv1.SeldonDeployment,
	p *machinelearningv1.PredictorSpec,
	addNamespace bool,
	serviceName string,
	serviceNameExternal string,
	customHeader string,
	customRegexHeader string,
	weight *int32,
	shadowing bool,
	engine_http_port int,
	isExplainer bool,
	instance_id string) (string, error) {

	namespace := utils2.GetNamespace(mlDep)

	// Set timeout
	timeout, err := strconv.Atoi(utils2.GetAnnotation(mlDep, ANNOTATION_REST_TIMEOUT, "3000"))
	if err != nil {
		return "", err
	}

	retries, err := strconv.Atoi(utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_RETRIES, AMBASSADOR_DEFAULT_RETRIES))
	if err != nil {
		return "", err
	}

	name := p.Name
	if isExplainer {
		name = p.Name + constants.ExplainerNameSuffix
		serviceNameExternal = serviceNameExternal + constants.ExplainerPathSuffix + "/" + p.Name
	}

	c := AmbassadorConfig{
		ApiVersion: "ambassador/v1",
		Kind:       "Mapping",
		Name:       "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_rest_mapping",
		Prefix:     "/seldon/" + serviceNameExternal + "/",
		Rewrite:    "/",
		Service:    serviceName + "." + namespace + ":" + strconv.Itoa(engine_http_port),
		TimeoutMs:  timeout,
	}

	// Ambassador only allows a single RetryOn: https://github.com/datawire/ambassador/issues/1570
	if retries != 0 {
		c.RetryPolicy = &AmbassadorRetryPolicy{
			RetryOn:    "gateway-error",
			NumRetries: retries,
		}
	}

	if weight != nil {
		c.Weight = *weight
	}

	if timeout > AMBASSADOR_IDLE_TIMEOUT {
		c.IdleTimeoutMs = &timeout
	}

	if addNamespace {
		c.Name = "seldon_" + namespace + "_" + mlDep.ObjectMeta.Name + "_" + name + "_rest_mapping"
		c.Prefix = "/seldon/" + namespace + "/" + serviceNameExternal + "/"
	}
	if customHeader != "" {
		headers := strings.Split(customHeader, ":")
		elementMap := make(map[string]string)
		for i := 0; i < len(headers); i += 2 {
			key := strings.TrimSpace(headers[i])
			val := strings.TrimSpace(headers[i+1])
			elementMap[key] = val
		}
		c.Headers = elementMap
	}
	if customRegexHeader != "" {
		headers := strings.Split(customHeader, ":")
		elementMap := make(map[string]string)
		for i := 0; i < len(headers); i += 2 {
			key := strings.TrimSpace(headers[i])
			val := strings.TrimSpace(headers[i+1])
			elementMap[key] = val
		}
		c.RegexHeaders = elementMap
	}
	if shadowing {
		c.Shadow = &shadowing
	}
	if instance_id != "" {
		c.InstanceId = instance_id
	}

	circuitBreakerConfig, err := getAmbassadorCircuitBreakerConfig(mlDep)
	if err != nil {
		return "", err
	}
	if circuitBreakerConfig != nil {
		c.CircuitBreakers = []*AmbassadorCircuitBreakerConfig{
			circuitBreakerConfig,
		}
	}

	if !utils.IsEmptyTLS(p) {
		if addNamespace {
			c.TLS = "seldon_" + namespace + "_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config"
		} else {
			c.TLS = "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config"
		}
	}

	v, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// Return a gRPC configuration for Ambassador with optional custom settings.
func getAmbassadorGrpcConfig(mlDep *machinelearningv1.SeldonDeployment,
	p *machinelearningv1.PredictorSpec,
	addNamespace bool,
	serviceName string,
	serviceNameExternal string,
	customHeader string,
	customRegexHeader string,
	weight *int32,
	shadowing bool,
	engine_grpc_port int,
	isExplainer bool,
	instance_id string) (string, error) {

	grpc := true
	namespace := utils2.GetNamespace(mlDep)

	// Set timeout
	timeout, err := strconv.Atoi(utils2.GetAnnotation(mlDep, ANNOTATION_GRPC_TIMEOUT, "3000"))
	if err != nil {
		return "", nil
	}

	retries, err := strconv.Atoi(utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_RETRIES, AMBASSADOR_DEFAULT_RETRIES))
	if err != nil {
		return "", err
	}

	name := p.Name
	if isExplainer {
		name = name + constants.ExplainerNameSuffix
		serviceNameExternal = serviceNameExternal + constants.ExplainerPathSuffix
	}

	c := AmbassadorConfig{
		ApiVersion:  "ambassador/v1",
		Kind:        "Mapping",
		Name:        "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_grpc_mapping",
		Grpc:        &grpc,
		Prefix:      constants.GRPCRegExMatchAmbassador,
		PrefixRegex: &grpc,
		Rewrite:     "",
		Headers:     map[string]string{"seldon": serviceNameExternal},
		Service:     serviceName + "." + namespace + ":" + strconv.Itoa(engine_grpc_port),
		TimeoutMs:   timeout,
	}

	// Ambassador only allows a single RetryOn: https://github.com/datawire/ambassador/issues/1570
	if retries != 0 {
		c.RetryPolicy = &AmbassadorRetryPolicy{
			RetryOn:    "gateway-error",
			NumRetries: retries,
		}
	}

	if weight != nil {
		c.Weight = *weight
	}

	if timeout > AMBASSADOR_IDLE_TIMEOUT {
		c.IdleTimeoutMs = &timeout
	}

	if addNamespace {
		c.Headers["namespace"] = namespace
		c.Name = "seldon_" + namespace + "_" + mlDep.ObjectMeta.Name + "_" + name + "_grpc_mapping"
	}
	if customHeader != "" {
		headers := strings.Split(customHeader, ":")
		for i := 0; i < len(headers); i += 2 {
			key := strings.TrimSpace(headers[i])
			val := strings.TrimSpace(headers[i+1])
			c.Headers[key] = val
		}
	}
	if customRegexHeader != "" {
		headers := strings.Split(customHeader, ":")
		elementMap := make(map[string]string)
		for i := 0; i < len(headers); i += 2 {
			key := strings.TrimSpace(headers[i])
			val := strings.TrimSpace(headers[i+1])
			elementMap[key] = val
		}
		c.RegexHeaders = elementMap
	}
	if shadowing {
		c.Shadow = &shadowing
	}
	if instance_id != "" {
		c.InstanceId = instance_id
	}

	circuitBreakerConfig, err := getAmbassadorCircuitBreakerConfig(mlDep)
	if err != nil {
		return "", err
	}
	if circuitBreakerConfig != nil {
		c.CircuitBreakers = []*AmbassadorCircuitBreakerConfig{
			circuitBreakerConfig,
		}
	}

	if !utils.IsEmptyTLS(p) {
		if addNamespace {
			c.TLS = "seldon_" + namespace + "_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config"
		} else {
			c.TLS = "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_tls_config"
		}
	}

	v, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// Get the configuration for ambassador using the servce name serviceName.
// Up to 4 confgurations will be created covering REST, GRPC and cluster-wide and namespaced varieties.
// Annotations for Ambassador will be used to customize the configuration returned.
func GetAmbassadorConfigs(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, serviceName string, engine_http_port, engine_grpc_port int, isExplainer bool) (string, error) {
	if annotation := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_CUSTOM, ""); annotation != "" {
		return annotation, nil
	} else {

		var weight *int32
		// Ignore weight on first predictor and let Ambassador handle this
		if mlDep.Spec.Predictors[0].Name != p.Name {
			weight = &p.Traffic
		}

		shadowing := p.Shadow
		serviceNameExternal := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_SERVICE, mlDep.GetName())
		customHeader := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_HEADER, "")
		customRegexHeader := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_REGEX_HEADER, "")
		instance_id := utils2.GetAnnotation(mlDep, ANNOTATION_AMBASSADOR_ID, "")

		cRestGlobal, err := getAmbassadorRestConfig(mlDep, p, true, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_http_port, isExplainer, instance_id)
		if err != nil {
			return "", err
		}
		cGrpcGlobal, err := getAmbassadorGrpcConfig(mlDep, p, true, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_grpc_port, isExplainer, instance_id)
		if err != nil {
			return "", err
		}
		cRestNamespaced, err := getAmbassadorRestConfig(mlDep, p, false, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_http_port, isExplainer, instance_id)
		if err != nil {
			return "", err
		}
		cGrpcNamespaced, err := getAmbassadorGrpcConfig(mlDep, p, false, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_grpc_port, isExplainer, instance_id)
		if err != nil {
			return "", err
		}
		cTLSGlobal, err := getAmbassadorTLSContextConfig(mlDep, p, false)
		if err != nil {
			return "", err
		}
		cTLSNamespaced, err := getAmbassadorTLSContextConfig(mlDep, p, true)
		if err != nil {
			return "", err
		}

		if utils.GetEnv("AMBASSADOR_SINGLE_NAMESPACE", "false") == "true" {
			return YAML_SEP + cRestGlobal + YAML_SEP + cGrpcGlobal + YAML_SEP + cTLSGlobal + YAML_SEP + cRestNamespaced + YAML_SEP + cGrpcNamespaced + YAML_SEP + cTLSNamespaced, nil
		} else {
			return YAML_SEP + cRestGlobal + YAML_SEP + cGrpcGlobal + YAML_SEP + cTLSGlobal, nil
		}
	}

}
