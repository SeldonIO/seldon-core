# Server Config

```{note}
This section is for advanced usage where you want to define new tyoes of inference servers.
```

Server configurations define how to create an inference server. By default one is provided for Seldon MLServer and one for NVIDIA Triton Inference Server. Both these servers support the V2 inference protocol which is a requirement for all inference servers. They define how the Kubernetes ReplicaSet is defined and which includes the Seldon Agent reverse proxy as well as an Rclone server for downloading artifacts for the server. The Kustomize ServerConfig for MlServer is shown below:

```{literalinclude} ../../../../../../operator/config/serverconfigs/mlserver.yaml
:language: yaml
```

