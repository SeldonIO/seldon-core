package controllers

import (
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

func getSvcOrchUser(mlDep *machinelearningv1.SeldonDeployment) (*int64, error) {

	if isExecutorEnabled(mlDep) {
		if envExecutorUser != "" {
			user, err := strconv.Atoi(envExecutorUser)
			if err != nil {
				return nil, err
			} else {
				engineUser := int64(user)
				return &engineUser, nil
			}
		}

	} else {
		if envEngineUser != "" {
			user, err := strconv.Atoi(envEngineUser)
			if err != nil {
				return nil, err
			} else {
				engineUser := int64(user)
				return &engineUser, nil
			}
		}
	}
	return nil, nil
}

func createSecurityContext(mlDep *machinelearningv1.SeldonDeployment) (*corev1.PodSecurityContext, error) {
	svcOrchUser, err := getSvcOrchUser(mlDep)
	if err != nil {
		return nil, err
	}
	if svcOrchUser != nil {
		return &corev1.PodSecurityContext{
			RunAsUser: svcOrchUser,
		}, nil
	}
	return nil, nil
}
