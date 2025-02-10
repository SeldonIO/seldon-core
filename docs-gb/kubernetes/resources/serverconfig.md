# Server Config

{% hint style="info" %}
**Note**: This section is for advanced usage where you want to define new types of inference servers.
{% endhint %}

Server configurations define how to create an inference server. By default one is provided
for Seldon MLServer and one for NVIDIA Triton Inference Server. Both these servers support
the V2 inference protocol which is a requirement for all inference servers. They define how
the Kubernetes ReplicaSet is defined which includes the Seldon Agent reverse proxy as well
as an Rclone server for downloading artifacts for the server. The Kustomize ServerConfig for
MlServer is shown below:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/operator/config/serverconfigs/mlserver.yaml" %}

