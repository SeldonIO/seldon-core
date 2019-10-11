package controllers

import (
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/api/v1alpha2"
	"gopkg.in/yaml.v2"
	"strconv"
	"strings"
)

const (
	ANNOTATION_REST_READ_TIMEOUT       = "seldon.io/rest-read-timeout"
	ANNOTATION_GRPC_READ_TIMEOUT       = "seldon.io/grpc-read-timeout"
	ANNOTATION_AMBASSADOR_CUSTOM       = "seldon.io/ambassador-config"
	ANNOTATION_AMBASSADOR_SHADOW       = "seldon.io/ambassador-shadow"
	ANNOTATION_AMBASSADOR_SERVICE      = "seldon.io/ambassador-service-name"
	ANNOTATION_AMBASSADOR_HEADER       = "seldon.io/ambassador-header"
	ANNOTATION_AMBASSADOR_REGEX_HEADER = "seldon.io/ambassador-regex-header"
	ANNOTATION_AMBASSADOR_ID           = "seldon.io/ambassador-id"

	YAML_SEP = "---\n"

	AMBASSADOR_IDLE_TIMEOUT = 300000
)

// Struct for Ambassador configuration
type AmbassadorConfig struct {
	ApiVersion    string                 `yaml:"apiVersion"`
	Kind          string                 `yaml:"kind"`
	Name          string                 `yaml:"name"`
	Grpc          *bool                  `yaml:"grpc,omitempty"`
	Prefix        string                 `yaml:"prefix"`
	Rewrite       string                 `yaml:"rewrite,omitempty"`
	Service       string                 `yaml:"service"`
	TimeoutMs     int                    `yaml:"timeout_ms"`
	IdleTimeoutMs *int                   `yaml:"idle_timeout_ms,omitempty"`
	Headers       map[string]string      `yaml:"headers,omitempty"`
	RegexHeaders  map[string]string      `yaml:"regex_headers,omitempty"`
	Weight        int32                  `yaml:"weight,omitempty"`
	Shadow        *bool                  `yaml:"shadow,omitempty"`
	RetryPolicy   *AmbassadorRetryPolicy `yaml:"retry_policy,omitempty"`
	InstanceId    string                 `yaml:"ambassador_id,omitempty"`
}

type AmbassadorRetryPolicy struct {
	RetryOn    string `yaml:"retry_on,omitempty"`
	NumRetries int    `yaml:"num_retries,omitempty"`
}

// Return a REST configuration for Ambassador with optional custom settings.
func getAmbassadorRestConfig(mlDep *machinelearningv1alpha2.SeldonDeployment,
	p *machinelearningv1alpha2.PredictorSpec,
	addNamespace bool,
	serviceName string,
	serviceNameExternal string,
	customHeader string,
	customRegexHeader string,
	weight int32,
	shadowing string,
	engine_http_port int,
	nameOverride string,
	instance_id string) (string, error) {

	namespace := getNamespace(mlDep)

	// Set timeout
	timeout, err := strconv.Atoi(getAnnotation(mlDep, ANNOTATION_REST_READ_TIMEOUT, "3000"))
	if err != nil {
		return "", nil
	}

	name := p.Name
	if nameOverride != "" {
		name = nameOverride
		serviceNameExternal = nameOverride
	}

	c := AmbassadorConfig{
		ApiVersion: "ambassador/v1",
		Kind:       "Mapping",
		Name:       "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_rest_mapping",
		Prefix:     "/seldon/" + serviceNameExternal + "/",
		Service:    serviceName + "." + namespace + ":" + strconv.Itoa(engine_http_port),
		TimeoutMs:  timeout,
		RetryPolicy: &AmbassadorRetryPolicy{
			RetryOn:    "connect-failure",
			NumRetries: 3,
		},
		Weight: weight,
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
	if shadowing != "" {
		shadow := true
		c.Shadow = &shadow
	}
	if instance_id != "" {
		c.InstanceId = instance_id
	}
	v, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// Return a gRPC configuration for Ambassador with optional custom settings.
func getAmbassadorGrpcConfig(mlDep *machinelearningv1alpha2.SeldonDeployment,
	p *machinelearningv1alpha2.PredictorSpec,
	addNamespace bool,
	serviceName string,
	serviceNameExternal string,
	customHeader string,
	customRegexHeader string,
	weight int32,
	shadowing string,
	engine_grpc_port int,
	nameOverride string,
	instance_id string) (string, error) {

	grpc := true
	namespace := getNamespace(mlDep)

	// Set timeout
	timeout, err := strconv.Atoi(getAnnotation(mlDep, ANNOTATION_GRPC_READ_TIMEOUT, "3000"))
	if err != nil {
		return "", nil
	}

	name := p.Name
	if nameOverride != "" {
		name = nameOverride
		serviceNameExternal = nameOverride
	}

	c := AmbassadorConfig{
		ApiVersion: "ambassador/v1",
		Kind:       "Mapping",
		Name:       "seldon_" + mlDep.ObjectMeta.Name + "_" + name + "_grpc_mapping",
		Grpc:       &grpc,
		Prefix:     "/seldon.protos.Seldon/",
		Rewrite:    "/seldon.protos.Seldon/",
		Headers:    map[string]string{"seldon": serviceNameExternal},
		Service:    serviceName + "." + namespace + ":" + strconv.Itoa(engine_grpc_port),
		TimeoutMs:  timeout,
		RetryPolicy: &AmbassadorRetryPolicy{
			RetryOn:    "connect-failure",
			NumRetries: 3,
		},
		Weight: weight,
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
	if shadowing != "" {
		shadow := true
		c.Shadow = &shadow
	}
	if instance_id != "" {
		c.InstanceId = instance_id
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
func getAmbassadorConfigs(mlDep *machinelearningv1alpha2.SeldonDeployment, p *machinelearningv1alpha2.PredictorSpec, serviceName string, engine_http_port, engine_grpc_port int, nameOverride string) (string, error) {
	if annotation := getAnnotation(mlDep, ANNOTATION_AMBASSADOR_CUSTOM, ""); annotation != "" {
		return annotation, nil
	} else {

		weight := p.Traffic
		if len(mlDep.Spec.Predictors) <= 1 {
			weight = 100
		}
		shadowing := getAnnotation(mlDep, ANNOTATION_AMBASSADOR_SHADOW, "")
		serviceNameExternal := getAnnotation(mlDep, ANNOTATION_AMBASSADOR_SERVICE, mlDep.ObjectMeta.Name)
		customHeader := getAnnotation(mlDep, ANNOTATION_AMBASSADOR_HEADER, "")
		customRegexHeader := getAnnotation(mlDep, ANNOTATION_AMBASSADOR_REGEX_HEADER, "")
		instance_id := getAnnotation(mlDep, ANNOTATION_AMBASSADOR_ID, "")

		cRestGlobal, err := getAmbassadorRestConfig(mlDep, p, true, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_http_port, nameOverride, instance_id)
		if err != nil {
			return "", err
		}
		cGrpcGlobal, err := getAmbassadorGrpcConfig(mlDep, p, true, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_grpc_port, nameOverride, instance_id)
		if err != nil {
			return "", err
		}
		cRestNamespaced, err := getAmbassadorRestConfig(mlDep, p, false, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_http_port, nameOverride, instance_id)
		if err != nil {
			return "", err
		}

		cGrpcNamespaced, err := getAmbassadorGrpcConfig(mlDep, p, false, serviceName, serviceNameExternal, customHeader, customRegexHeader, weight, shadowing, engine_grpc_port, nameOverride, instance_id)
		if err != nil {
			return "", err
		}

		// Return the appropriate set of config based on whether http and/or grpc is active
		if engine_http_port > 0 && engine_grpc_port > 0 {
			if GetEnv("AMBASSADOR_SINGLE_NAMESPACE", "false") == "true" {
				return YAML_SEP + cRestGlobal + YAML_SEP + cGrpcGlobal + YAML_SEP + cRestNamespaced + YAML_SEP + cGrpcNamespaced, nil
			} else {
				return YAML_SEP + cRestGlobal + YAML_SEP + cGrpcGlobal, nil
			}
		} else if engine_http_port > 0 {
			if GetEnv("AMBASSADOR_SINGLE_NAMESPACE", "false") == "true" {
				return YAML_SEP + cRestGlobal + YAML_SEP + cRestNamespaced, nil
			} else {
				return YAML_SEP + cRestGlobal, nil
			}
		} else if engine_grpc_port > 0 {
			if GetEnv("AMBASSADOR_SINGLE_NAMESPACE", "false") == "true" {
				return YAML_SEP + cGrpcGlobal + YAML_SEP + cGrpcNamespaced, nil
			} else {
				return YAML_SEP + cGrpcGlobal, nil
			}
		} else {
			return "", nil
		}

	}

}
