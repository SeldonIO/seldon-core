# Creation

This operator was built with  [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

```
Version: main.version{KubeBuilderVersion:"3.2.0", KubernetesVendor:"1.22.1", GitCommit:"b7a730c84495122a14a0faff95e9e9615fffbfc5", BuildDate:"2021-10-29T18:32:16Z", GoOs:"linux", GoArch:"amd64"}
```

Steps to recreate scaffolding

```
go mod init github.com/seldonio/seldon-core/operatorv2
kubebuilder init --domain seldon.io
kubebuilder edit --multigroup=true
kubebuilder create api --group mlops --version v1alpha1 --kind Model --resource --controller
kubebuilder create api --group mlops --version v1alpha1 --kind Server --resource --controller
kubebuilder create api --group mlops --version v1alpha1 --kind Pipeline --resource --controller
kubebuilder create api --group mlops --version v1alpha1 --kind Explainer --resource --controller
```

