package controllers

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var nilRunAsUid *int64 = nil
var runAsUid int64 = 1337

func TestStorageInitalizerInjector(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme = createScheme()
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().ConfigMaps(ControllerNamespace).Create(context.TODO(), configMap, v1meta.CreateOptions{})
	g.Expect(err).To(BeNil())
	mi := NewModelInitializer(context.TODO(), client)
	containerName := "classifier"
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: containerName,
						},
					},
				},
			},
		},
	}
	_, err = mi.InjectModelInitializer(&d, containerName, "gs://mybucket/mymodel", "", "", "", &runAsUid)
	g.Expect(err).To(BeNil())
	g.Expect(len(d.Spec.Template.Spec.InitContainers)).To(Equal(1))
	g.Expect(d.Spec.Template.Spec.InitContainers[0].Image).To(Equal("kfserving/storage-initializer:v0.6.1"))
	g.Expect(*d.Spec.Template.Spec.InitContainers[0].SecurityContext.RunAsUser).To(Equal(runAsUid))
}

func TestStorageInitalizerInjectorWithRelatedImage(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme = createScheme()
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().ConfigMaps(ControllerNamespace).Create(context.TODO(), configMap, v1meta.CreateOptions{})
	g.Expect(err).To(BeNil())
	mi := NewModelInitializer(context.TODO(), client)
	containerName := "classifier"
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: containerName,
						},
					},
				},
			},
		},
	}
	envStorageInitializerImage = "abc:1.2"
	_, err = mi.InjectModelInitializer(&d, containerName, "gs://mybucket/mymodel", "", "", "", nilRunAsUid)
	g.Expect(err).To(BeNil())
	g.Expect(len(d.Spec.Template.Spec.InitContainers)).To(Equal(1))
	g.Expect(d.Spec.Template.Spec.InitContainers[0].Image).To(Equal(envStorageInitializerImage))
	envStorageInitializerImage = ""
}

func TestStorageInitalizerInjectorWithGraphDefinedImage(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme = createScheme()
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().ConfigMaps(ControllerNamespace).Create(context.TODO(), configMap, v1meta.CreateOptions{})
	g.Expect(err).To(BeNil())
	mi := NewModelInitializer(context.TODO(), client)
	containerName := "classifier"
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: containerName,
						},
					},
				},
			},
		},
	}
	storageInitializerImage := "abc:1.3"
	_, err = mi.InjectModelInitializer(&d, containerName, "gs://mybucket/mymodel", "", "", storageInitializerImage, nilRunAsUid)
	g.Expect(err).To(BeNil())
	g.Expect(len(d.Spec.Template.Spec.InitContainers)).To(Equal(1))
	g.Expect(d.Spec.Template.Spec.InitContainers[0].Image).To(Equal(storageInitializerImage))
	g.Expect(d.Spec.Template.Spec.InitContainers[0].SecurityContext).To(BeNil())
}

func TestStorageInitalizerInjectorWithGraphDefinedImagePriorityOverRelated(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme = createScheme()
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().ConfigMaps(ControllerNamespace).Create(context.TODO(), configMap, v1meta.CreateOptions{})
	g.Expect(err).To(BeNil())
	mi := NewModelInitializer(context.TODO(), client)
	containerName := "classifier"
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: containerName,
						},
					},
				},
			},
		},
	}
	envStorageInitializerImage = "abc:1.2"
	storageInitializerImage := "abc:1.3"
	_, err = mi.InjectModelInitializer(&d, containerName, "gs://mybucket/mymodel", "", "", storageInitializerImage, nilRunAsUid)
	g.Expect(err).To(BeNil())
	g.Expect(len(d.Spec.Template.Spec.InitContainers)).To(Equal(1))
	g.Expect(d.Spec.Template.Spec.InitContainers[0].Image).To(Equal(storageInitializerImage))
	envStorageInitializerImage = ""
}
