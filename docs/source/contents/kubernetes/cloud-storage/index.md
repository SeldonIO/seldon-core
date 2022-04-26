# Cloud Storage

Inference artifacts referenced in Models can be stored in local mounted folders or on any of the cloud storage technologies supported by [Rclone](https://rclone.org/). Public Google buckets work by default which allows us to use examples such as below:

```{literalinclude} ../../../../../samples/models/sklearn-iris-gs.yaml 
:language: yaml
```

The format is described [here](https://rclone.org/rc/#config-create). The main requirements will be to choose a particular `type` and `name` to use in storage urls and set the parameters as described in the Rclone docs.

To add authorization for cloud storage you need to define an Rclone provider in one of three ways:

## Kubernetes Secret

You can provide the provider credentials in a Kubernetes secret. For example, assuming minio has be installed in the cluster an example secret would be:

```{literalinclude} ../../../../../samples/auth/minio-secret.yaml
:language: yaml
```

Yoiu can then reference this in a Model, such as:

```{literalinclude} ../../../../../samples/models/sklearn-iris-minio.yaml
:language: yaml
```


## Central Config Map

To allow all models to utilize particular rclone providers one can add the secrets to the agent configMap, e.g.

```{literalinclude} ../../../../../samples/auth/agent.yaml
:language: yaml
```


