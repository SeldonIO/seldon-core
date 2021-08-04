/*
Copyright 2019 kubeflow.org.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s3

import (
	"k8s.io/api/core/v1"
)

const (
	AWSAccessKeyId         = "AWS_ACCESS_KEY_ID"
	AWSSecretAccessKey     = "AWS_SECRET_ACCESS_KEY"
	AWSAccessKeyIdName     = "awsAccessKeyID"
	AWSSecretAccessKeyName = "awsSecretAccessKey"
	AWSEndpointUrl         = "AWS_ENDPOINT_URL"
	AWSRegion              = "AWS_REGION"
	S3Endpoint             = "S3_ENDPOINT"
	S3UseHttps             = "S3_USE_HTTPS"
	S3VerifySSL            = "S3_VERIFY_SSL"
	KFServingAPIGroupName  = "serving.kubeflow.org"
	SeldonAPIGroupName     = "machinelearning.seldon.io"
)

type S3Config struct {
	S3AccessKeyIDName     string `json:"s3AccessKeyIDName,omitempty"`
	S3SecretAccessKeyName string `json:"s3SecretAccessKeyName,omitempty"`
	S3Endpoint            string `json:"s3Endpoint,omitempty"`
	S3UseHttps            string `json:"s3UseHttps,omitempty"`
}

// prefix to annotation could be SeldonAPIGroupName or KFServingAPIGroupName
var (
	S3SecretEndpointAnnotation = "/" + "s3-endpoint"
	S3SecretRegionAnnotation   = "/" + "s3-region"
	S3SecretSSLAnnotation      = "/" + "s3-verifyssl"
	S3SecretHttpsAnnotation    = "/" + "s3-usehttps"
)

func BuildSecretEnvs(secret *v1.Secret, s3Config *S3Config) []v1.EnvVar {
	s3SecretAccessKeyName := AWSSecretAccessKeyName
	s3AccessKeyIdName := AWSAccessKeyIdName
	if s3Config.S3AccessKeyIDName != "" {
		s3AccessKeyIdName = s3Config.S3AccessKeyIDName
	}

	if s3Config.S3SecretAccessKeyName != "" {
		s3SecretAccessKeyName = s3Config.S3SecretAccessKeyName
	}
	envs := []v1.EnvVar{
		{
			Name: AWSAccessKeyId,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: s3AccessKeyIdName,
				},
			},
		},
		{
			Name: AWSSecretAccessKey,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: s3SecretAccessKeyName,
				},
			},
		},
	}
	envsFromAnnotations := BuildEnvFromAnnotations(secret, s3Config, SeldonAPIGroupName, KFServingAPIGroupName)
	envs = append(envs, envsFromAnnotations...)
	return envs
}

func getFromAnnotation(secret *v1.Secret, name string, prefix string, fallbackPrefix string) (string, bool) {
	if value, ok := secret.Annotations[prefix+name]; ok {
		return value, ok
	} else if value, ok := secret.Annotations[fallbackPrefix+name]; ok {
		return value, ok
	}
	return "", false
}

func BuildEnvFromAnnotations(secret *v1.Secret, s3Config *S3Config, prefix string, fallbackPrefix string) []v1.EnvVar {
	envs := []v1.EnvVar{}

	if s3Endpoint, ok := getFromAnnotation(secret, S3SecretEndpointAnnotation, prefix, fallbackPrefix); ok {
		s3EndpointUrl := "https://" + s3Endpoint
		if s3UseHttps, ok := getFromAnnotation(secret, S3SecretHttpsAnnotation, prefix, fallbackPrefix); ok {
			if s3UseHttps == "0" {
				s3EndpointUrl, _ = getFromAnnotation(secret, S3SecretEndpointAnnotation, prefix, fallbackPrefix)
				s3EndpointUrl = "http://" + s3EndpointUrl
			}
			envs = append(envs, v1.EnvVar{
				Name:  S3UseHttps,
				Value: s3UseHttps,
			})
		}
		envs = append(envs, v1.EnvVar{
			Name:  S3Endpoint,
			Value: s3Endpoint,
		})
		envs = append(envs, v1.EnvVar{
			Name:  AWSEndpointUrl,
			Value: s3EndpointUrl,
		})
	} else if s3Config.S3Endpoint != "" {
		s3EndpointUrl := "https://" + s3Config.S3Endpoint
		if s3Config.S3UseHttps == "0" {
			s3EndpointUrl = "http://" + s3Config.S3Endpoint
			envs = append(envs, v1.EnvVar{
				Name:  S3UseHttps,
				Value: s3Config.S3UseHttps,
			})
		}
		envs = append(envs, v1.EnvVar{
			Name:  S3Endpoint,
			Value: s3Config.S3Endpoint,
		})
		envs = append(envs, v1.EnvVar{
			Name:  AWSEndpointUrl,
			Value: s3EndpointUrl,
		})
	}

	if s3Region, ok := getFromAnnotation(secret, S3SecretRegionAnnotation, prefix, fallbackPrefix); ok {
		envs = append(envs, v1.EnvVar{
			Name:  AWSRegion,
			Value: s3Region,
		})
	}

	if val, ok := getFromAnnotation(secret, S3SecretSSLAnnotation, prefix, fallbackPrefix); ok {
		envs = append(envs, v1.EnvVar{
			Name:  S3VerifySSL,
			Value: val,
		})
	}
	return envs
}
