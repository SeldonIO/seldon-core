# Seldon Core Release 0.2.7

A summary of the main contributions to the [Seldon Core release 0.2.7](https://github.com/SeldonIO/seldon-core/releases/tag/v0.2.7).

## Autoscaling
We now provide the ability to autoscale your Seldon deployments using Kubernetes Horizontal Pod Autoscalers (HPA). When you define your Seldon Deployment you can add an HPA spec for each predictor. As each predictor in Seldon is exposed as a Kubernetes Deployment the HPA spec is attached to that deployment. An example SeldonDeployment is shown below:

```json
{
  "apiVersion": "machinelearning.seldon.io/v1alpha2",
  "kind": "SeldonDeployment",
  "metadata": {
    "name": "seldon-model"
  },
  "spec": {
    "name": "test-deployment",
    "oauth_key": "oauth-key",
    "oauth_secret": "oauth-secret",
    "predictors": [
      {
        "componentSpecs": [
          {
            "spec": {
              "containers": [
                {
                  "image": "seldonio/mock_classifier:1.0",
                  "imagePullPolicy": "IfNotPresent",
                  "name": "classifier",
                  "resources": {
                    "requests": {
                      "cpu": "0.5"
                    }
                  }
                }
              ],
              "terminationGracePeriodSeconds": 1
            },
            "hpaSpec": {
              "minReplicas": 1,
              "maxReplicas": 4,
              "metrics": [
                {
                  "type": "Resource",
                  "resource": {
                    "name": "cpu",
                    "targetAverageUtilization": 10
                  }
                }
              ]
            }
          }
        ],
        "graph": {
          "children": [],
          "name": "classifier",
          "endpoint": {
            "type": "REST"
          },
          "type": "MODEL"
        },
        "name": "example",
        "replicas": 1
      }
    ]
  }
}
```

In the above we can see, we added resource requests for the cpu:

```json
{
  "resources": {
    "requests": {
      "cpu": "0.5"
    }
  }
}
```

We added an HPA spec referring to the PodTemplateSpec:

```json
{
  "hpaSpecs": [
    {
      "minReplicas": 1,
      "maxReplicas": 4,
      "metrics": [
        {
          "type": "Resource",
          "resource": {
            "name": "cpu",
            "targetAverageUtilization": 10
          }
        }
      ]
    }
  ]
}
```

For full documentation see [here](../graph/scaling.html)

## Official Ambassador and Redis Helm Charts
We have updated our `seldon-core` helm chart to utilize the official Helm charts for [Ambassador](https://github.com/helm/charts/tree/master/stable/ambassador) and [Redis](https://github.com/helm/charts/tree/master/stable/redis). This provides users with a clear installation artifact for these components and also provides the flexibility for installing these components standalone if desired. It is planned in future releases to remove the particular ingress component from our Helm charts and for users to install these themselves for the API gateway they wish to use, for example Ambassador, Seldon's OAuth gateway or in future [istio-ingress gateway](https://istio.io/docs/tasks/traffic-management/ingress/).

## Developer Prow Integration
We have integrated [Prow](https://github.com/kubernetes/test-infra/tree/master/prow) into our development process. Prow is a Kubernetes based CI/CD system. At present it does some basic automatic labelling of pull requests but we plan to expand its usage in future.

