# Creation

This operator was built with  [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

```
kubebuilder version
Version: main.version{KubeBuilderVersion:"3.1.0", KubernetesVendor:"1.19.2", GitCommit:"92e0349ca7334a0a8e5e499da4fb077eb524e94a", BuildDate:"2021-05-27T17:54:28Z", GoOs:"linux", GoArch:"amd64"}
```

Steps to recreate scaffolding

```
go mod init github.com/seldonio/seldon-core/operatorv2
kubebuilder init --domain seldon.io
kubebuilder edit --multigroup=true
kubebuilder create api --group mlops --version v1alpha1 --kind InferenceArtifact --resource --controller
kubebuilder create api --group mlops --version v1alpha1 --kind InferenceServer --resource --controller
kubebuilder create api --group mlops --version v1alpha1 --kind InferenceGraph --resource --controller
kubebuilder create api --group mlops --version v1alpha1 --kind InferenceExplainer --resource --controller
kubebuilder create api --group mlops --version v1alpha1 --kind InferenceServerInstance --resource --controller
```

