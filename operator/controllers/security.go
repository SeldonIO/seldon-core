package controllers

import (
	"os"
	"strconv"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	ENV_DEFAULT_USER_ID = "DEFAULT_USER_ID"
)

var (
	envDefaultUser = os.Getenv(ENV_DEFAULT_USER_ID)
)

func createSecurityContext(mlDep *machinelearningv1.SeldonDeployment) (*corev1.PodSecurityContext, error) {
	if envDefaultUser != "" {
		user, err := strconv.Atoi(envDefaultUser)
		if err != nil {
			return nil, err
		}
		userId := int64(user)
		return &corev1.PodSecurityContext{
			RunAsUser: &userId,
		}, nil
	}
	return nil, nil
}
