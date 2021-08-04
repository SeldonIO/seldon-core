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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"os"

	"github.com/seldonio/seldon-core/operator/controllers/resources/credentials"
	"github.com/seldonio/seldon-core/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TODO: change image to seldon? is at least configurable by configmap now (with fixed version there)
// TODO: check PVC
const (
	DefaultModelLocalMountPath         = "/mnt/models"
	StorageInitializerConfigMapKeyName = "storageInitializer"
	ModelInitializerContainerImage     = "gcr.io/kfserving/model-initializer"
	ModelInitializerContainerVersion   = "latest"
	PvcURIPrefix                       = "pvc://"
	PvcSourceMountName                 = "kfserving-pvc-source"
	PvcSourceMountPath                 = "/mnt/pvc"
	ModelInitializerVolumeSuffix       = "provision-location"
	ModelInitializerContainerSuffix    = "model-initializer"
	EnvStorageInitializerImageRelated  = "RELATED_IMAGE_STORAGE_INITIALIZER"
)

var (
	ControllerNamespace        = utils.GetEnv("POD_NAMESPACE", "seldon-system")
	ControllerConfigMapName    = "seldon-config"
	envStorageInitializerImage = os.Getenv(EnvStorageInitializerImageRelated)
)

type ModelInitialiser struct {
	clientset kubernetes.Interface
	ctx       context.Context
}

func NewModelInitializer(ctx context.Context, clientset kubernetes.Interface) *ModelInitialiser {
	return &ModelInitialiser{clientset: clientset, ctx: ctx}
}

type StorageInitializerConfig struct {
	Image         string `json:"image"`
	CpuRequest    string `json:"cpuRequest"`
	CpuLimit      string `json:"cpuLimit"`
	MemoryRequest string `json:"memoryRequest"`
	MemoryLimit   string `json:"memoryLimit"`
}

func (mi *ModelInitialiser) credentialsBuilder() (credentialsBuilder *credentials.CredentialBuilder, err error) {
	configMap, err := mi.clientset.CoreV1().ConfigMaps(ControllerNamespace).Get(mi.ctx, ControllerConfigMapName, metav1.GetOptions{})
	if err != nil {
		//log.Error(err, "Failed to find config map", "name", ControllerConfigMapName)
		return nil, err
	}

	credentialBuilder := credentials.NewCredentialBulder(mi.ctx, configMap, mi.clientset)
	return credentialBuilder, nil
}

func (mi *ModelInitialiser) getStorageInitializerConfigs() (*StorageInitializerConfig, error) {
	configMap, err := mi.clientset.CoreV1().ConfigMaps(ControllerNamespace).Get(mi.ctx, ControllerConfigMapName, metav1.GetOptions{})
	if err != nil {
		//log.Error(err, "Failed to find config map", "name", ControllerConfigMapName)
		return nil, err
	}
	return getStorageInitializerConfigsFromMap(configMap)
}

func getStorageInitializerConfigsFromMap(configMap *corev1.ConfigMap) (*StorageInitializerConfig, error) {
	storageInitializerConfig := &StorageInitializerConfig{}
	if initializerConfig, ok := configMap.Data[StorageInitializerConfigMapKeyName]; ok {
		err := json.Unmarshal([]byte(initializerConfig), &storageInitializerConfig)
		if err != nil {
			panic(fmt.Errorf("Unable to unmarshall %v json string due to %v ", StorageInitializerConfigMapKeyName, err))
		}
	}
	//Ensure that we set proper values for CPU/Memory Limit/Request
	resourceDefaults := []string{storageInitializerConfig.MemoryRequest,
		storageInitializerConfig.MemoryLimit,
		storageInitializerConfig.CpuRequest,
		storageInitializerConfig.CpuLimit}
	for _, key := range resourceDefaults {
		_, err := resource.ParseQuantity(key)
		if err != nil {
			return storageInitializerConfig, err
		}
	}

	return storageInitializerConfig, nil
}

// InjectModelInitializer injects an init container to provision model data
func (mi *ModelInitialiser) InjectModelInitializer(deployment *appsv1.Deployment, containerName string, srcURI string, serviceAccountName string, envSecretRefName string, storageInitializerImage string) (deploy *appsv1.Deployment, err error) {

	if srcURI == "" {
		return deployment, nil
	}

	userContainer := utils.GetContainerForDeployment(deployment, containerName)

	if userContainer == nil {
		return deployment, fmt.Errorf("Invalid configuration: cannot find container: %s", containerName)
	}

	ModelInitializerVolumeName := userContainer.Name + "-" + ModelInitializerVolumeSuffix
	//kubernetes names limited to 63
	if len(ModelInitializerVolumeName) > 63 {
		ModelInitializerVolumeName = ModelInitializerVolumeName[0:63]
		ModelInitializerVolumeName = strings.TrimSuffix(ModelInitializerVolumeName, "-")
	}

	ModelInitializerContainerName := userContainer.Name + "-" + ModelInitializerContainerSuffix
	if len(ModelInitializerContainerName) > 63 {
		ModelInitializerContainerName = ModelInitializerContainerName[0:63]
		ModelInitializerContainerName = strings.TrimSuffix(ModelInitializerContainerName, "-")
	}

	// TODO: KFServing does a check for an annotation before injecting - not doing that for now
	podSpec := &deployment.Spec.Template.Spec

	// Dont inject if InitContianer already injected
	for _, container := range podSpec.InitContainers {
		if strings.Compare(container.Name, ModelInitializerContainerName) == 0 {
			// make sure we have the mount on the main container
			addVolumeMountToContainer(userContainer, ModelInitializerVolumeName, DefaultModelLocalMountPath)
			return deployment, nil
		}
	}

	podVolumes := []corev1.Volume{}
	modelInitializerMounts := []corev1.VolumeMount{}

	// For PVC source URIs we need to mount the source to be able to access it
	// See design and discussion here: https://github.com/kubeflow/kfserving/issues/148
	if strings.HasPrefix(srcURI, PvcURIPrefix) {
		pvcName, pvcPath, err := parsePvcURI(srcURI)
		if err != nil {
			return nil, err
		}

		// add the PVC volume on the pod
		pvcSourceVolume := corev1.Volume{
			Name: PvcSourceMountName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		}
		podVolumes = append(podVolumes, pvcSourceVolume)

		// add a corresponding PVC volume mount to the INIT container
		pvcSourceVolumeMount := corev1.VolumeMount{
			Name:      PvcSourceMountName,
			MountPath: PvcSourceMountPath,
			ReadOnly:  true,
		}
		modelInitializerMounts = append(modelInitializerMounts, pvcSourceVolumeMount)

		// Since the model path is linked from source pvc, userContainer also need to mount the pvc.
		addVolumeMountToContainer(userContainer, PvcSourceMountName, PvcSourceMountPath)

		// modify the sourceURI to point to the PVC path
		srcURI = PvcSourceMountPath + "/" + pvcPath
	}

	// Create a volume that is shared between the model-initializer and user-container
	sharedVolume := corev1.Volume{
		Name: ModelInitializerVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	podVolumes = append(podVolumes, sharedVolume)

	// Create a write mount into the shared volume
	sharedVolumeWriteMount := corev1.VolumeMount{
		Name:      ModelInitializerVolumeName,
		MountPath: DefaultModelLocalMountPath,
		ReadOnly:  false,
	}
	modelInitializerMounts = append(modelInitializerMounts, sharedVolumeWriteMount)

	config, err := mi.getStorageInitializerConfigs()
	if err != nil {
		return nil, err
	}

	initContainerImage := ModelInitializerContainerImage + ":" + ModelInitializerContainerVersion

	if storageInitializerImage != "" {
		initContainerImage = storageInitializerImage
	} else if envStorageInitializerImage != "" {
		initContainerImage = envStorageInitializerImage
	} else if config != nil && config.Image != "" {
		initContainerImage = config.Image
	}

	// Add an init container to run provisioning logic to the PodSpec (with defaults to pass comparison later)
	initContainer := &corev1.Container{
		Name:            ModelInitializerContainerName,
		Image:           initContainerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args: []string{
			srcURI,
			DefaultModelLocalMountPath,
		},
		VolumeMounts:             modelInitializerMounts,
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		Resources: corev1.ResourceRequirements{
			Limits: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse(config.CpuLimit),
				corev1.ResourceMemory: resource.MustParse(config.MemoryLimit),
			},
			Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse(config.CpuRequest),
				corev1.ResourceMemory: resource.MustParse(config.MemoryRequest),
			},
		},
	}

	addVolumeMountToContainer(userContainer, ModelInitializerVolumeName, DefaultModelLocalMountPath)
	// Add volumes to the PodSpec
	podSpec.Volumes = append(podSpec.Volumes, podVolumes...)

	// Inject credentials
	credentialsBuilder, err := mi.credentialsBuilder()
	if err != nil {
		return nil, err
	}
	if serviceAccountName == "" {
		serviceAccountName = podSpec.ServiceAccountName
	}

	if err := credentialsBuilder.CreateSecretVolumeAndEnv(
		deployment.Namespace,
		serviceAccountName,
		initContainer,
		&podSpec.Volumes,
	); err != nil {
		return nil, err
	}

	// Inject credentials using secretRef
	if envSecretRefName != "" {
		initContainer.EnvFrom = append(initContainer.EnvFrom,
			corev1.EnvFromSource{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: envSecretRefName,
					},
				},
			})
	}

	// Add init container to the spec
	podSpec.InitContainers = append(podSpec.InitContainers, *initContainer)

	return deployment, nil
}

func addVolumeMountToContainer(userContainer *corev1.Container, ModelInitializerVolumeName string, MountPath string) {
	mountFound := false
	for _, v := range userContainer.VolumeMounts {
		if v.Name == ModelInitializerVolumeName {
			mountFound = true
		}
	}
	if !mountFound {
		// Add a mount the shared volume on the user-container, update the PodSpec
		sharedVolumeReadMount := &corev1.VolumeMount{
			Name:      ModelInitializerVolumeName,
			MountPath: MountPath,
			ReadOnly:  true,
		}
		userContainer.VolumeMounts = append(userContainer.VolumeMounts, *sharedVolumeReadMount)
	}
}

func parsePvcURI(srcURI string) (pvcName string, pvcPath string, err error) {
	parts := strings.Split(strings.TrimPrefix(srcURI, PvcURIPrefix), "/")
	if len(parts) > 1 {
		pvcName = parts[0]
		pvcPath = strings.Join(parts[1:], "/")
	} else if len(parts) == 1 {
		pvcName = parts[0]
		pvcPath = ""
	} else {
		return "", "", fmt.Errorf("Invalid URI must be pvc://<pvcname>/[path]: %s", srcURI)
	}

	return pvcName, pvcPath, nil
}
