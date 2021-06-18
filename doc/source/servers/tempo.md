# Tempo Server

[Tempo](https://github.com/SeldonIO/tempo) is a MLOps python SDK that allows packaging custom python servers and orchestration of multiple models from python. The Tempo python SDK allows packaging of the custom code as a conda-pack environment tar ball and Cloudpickle artifacts. It has a Seldon Core runtime which allows Tempo artifacts to be run under Seldon Core.

For more details see the [Tempo documentation](https://tempo.readthedocs.io/en/latest/).

An example Tempo model yaml for Seldon Core is shown below:

```python
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  annotations:
    seldon.io/tempo-description: ''
    seldon.io/tempo-model: '{"model_details": {"name": "numpyro-divorce", "local_folder":
      "/home/clive/work/mlops/fork-tempo/docs/examples/custom-model/artifacts", "uri":
      "s3://tempo/divorce", "platform": "custom", "inputs": {"args": [{"ty": "numpy.ndarray",
      "name": "marriage"}, {"ty": "numpy.ndarray", "name": "age"}]}, "outputs": {"args":
      [{"ty": "numpy.ndarray", "name": null}]}, "description": ""}, "protocol": "tempo.kfserving.protocol.KFServingV2Protocol",
      "runtime_options": {"runtime": "tempo.seldon.SeldonKubernetesRuntime", "docker_options":
      {"defaultRuntime": "tempo.seldon.SeldonDockerRuntime"}, "k8s_options": {"replicas":
      1, "minReplicas": null, "maxReplicas": null, "authSecretName": "minio-secret",
      "serviceAccountName": null, "defaultRuntime": "tempo.seldon.SeldonKubernetesRuntime",
      "namespace": "production"}, "ingress_options": {"ingress": "tempo.ingress.istio.IstioIngress",
      "ssl": false, "verify_ssl": true}}}'
  labels:
    seldon.io/tempo: 'true'
  name: numpyro-divorce
  namespace: production
spec:
  predictors:
  - graph:
      envSecretRefName: minio-secret
      implementation: TEMPO_SERVER
      modelUri: s3://tempo/divorce
      name: numpyro-divorce
      serviceAccountName: tempo-pipeline
      type: MODEL
    name: default
    replicas: 1
  protocol: kfserving
```
