package utils

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetEnvAsBool(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		raw      string
		expected bool
	}{
		{
			raw:      "true",
			expected: true,
		},
		{
			raw:      "TRUE",
			expected: true,
		},
		{
			raw:      "1",
			expected: true,
		},
		{
			raw:      "false",
			expected: false,
		},
		{
			raw:      "FALSE",
			expected: false,
		},
		{
			raw:      "0",
			expected: false,
		},
		{
			raw:      "foo",
			expected: false,
		},
		{
			raw:      "",
			expected: false,
		},
		{
			raw:      "345",
			expected: false,
		},
	}

	for _, test := range tests {
		os.Setenv("TEST_FOO", test.raw)
		val := GetEnvAsBool("TEST_FOO", false)

		g.Expect(val).To(Equal(test.expected))
	}
}

func TestAddEnvVarToDeploymentContainers(t *testing.T) {

	g := NewGomegaWithT(t)

	container := corev1.Container{}
	testName := "test-name"
	testValue := "test-value"

	deploy := &appsv1.Deployment{}
	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, container)

	envVar := &corev1.EnvVar{Name: testName, Value: testValue}
	AddEnvVarToDeploymentContainers(deploy, envVar)

	container = deploy.Spec.Template.Spec.Containers[0]
	g.Expect(len(container.Env)).To(Equal(1))
	g.Expect((container).Env[0].Value).To(Equal(testValue))

	testValueModified := "test-value-modified"
	envVarModified := &corev1.EnvVar{Name: testName, Value: testValueModified}
	AddEnvVarToDeploymentContainers(deploy, envVarModified)

	// We expect to still be the same and unmodified as it should not override
	container = deploy.Spec.Template.Spec.Containers[0]
	g.Expect(len(container.Env)).To(Equal(1))
	g.Expect(container.Env[0].Value).To(Equal(testValue))
}

func TestMountSecretToDeploymentContainers(t *testing.T) {

	g := NewGomegaWithT(t)

	testSecretRefName := "secret-name"
	testContainerMountPath := "/container/mount/path"

	container := corev1.Container{}

	deploy := &appsv1.Deployment{}
	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, container)

	MountSecretToDeploymentContainers(deploy, testSecretRefName, testContainerMountPath)

	volumes := &deploy.Spec.Template.Spec.Volumes
	g.Expect(len(*volumes)).To(Equal(1))
	g.Expect((*volumes)[0].VolumeSource.Secret.SecretName).To(Equal(testSecretRefName))
	volumeMounts := &deploy.Spec.Template.Spec.Containers[0].VolumeMounts
	g.Expect(len(*volumeMounts)).To(Equal(1))
	g.Expect((*volumeMounts)[0].MountPath).To(Equal(testContainerMountPath))
}
