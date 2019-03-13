# Pulling from Private Docker Registries

To pull images from private Docker registries simply add imagePullSecrets to the podTemplateSpecs for your SeldonDeployment resources. For example, show below is a simple model which uses a private image ```private-docker-repo/my-image```.  You will need to have created the Kubernetes docker registry secret ```myreposecret``` before applying the resource to your cluster.

```
{
  apiVersion: "machinelearning.seldon.io/v1alpha2",
  kind: "SeldonDeployment",
  metadata: {
    name: private-model,
  },
  spec: {
    name: private-model-example,
    predictors: [
      {
        componentSpecs: [{
          spec: {
            containers: [
              {
                image: private-docker-repo/my-image,
                name: private-model,
              },
            ],
	    imagePullSecrets: [
              {
                name: myreposecret
              },
            ],
          },
        }],
        graph: {
          children: [],
          endpoint: {
            type: REST,
          },
          name: private-model,
          type: "MODEL",
        },
        name: private-model,
        replicas: 1,
      },
    ],
  },
}
```

To create the docker registry secret see the [Kubernetes docs](https://kubernetes.io/docs/concepts/containers/images/#creating-a-secret-with-a-docker-config).
