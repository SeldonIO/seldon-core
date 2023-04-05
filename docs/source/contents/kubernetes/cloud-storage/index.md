# Cloud Storage

Inference artifacts referenced in Models can be stored in local mounted folders or on any of the cloud storage technologies supported by [Rclone](https://rclone.org/).
Public Google buckets work by default which allows us to use examples such as below:

```{literalinclude} ../../../../../samples/models/sklearn-iris-gs.yaml 
:language: yaml
```

The format for defining your Rclone storage credentials is described [here](https://rclone.org/rc/#config-create).
The main requirements will be to choose a particular `type` and `name` to use in storage urls and set the parameters as described in the Rclone docs where the parameters follow the given options described in the docs where for example `--gcs-client-secret` can be added as a paramater `client_secret`, i.e. without the type prefix and with underscores.

To add authorization for cloud storage you need to define an Rclone provider as discussed below in a Kubernetes Secret.

## Kubernetes Secret

You can provide the provider credentials in a Kubernetes secret.
We show some examples in the following sections.

### S3 Minio Example

Assuming minio has be installed in the cluster an example secret would be:

```{literalinclude} ../../../../../samples/auth/minio-secret.yaml
:language: yaml
```

Yoiu can then reference this in a Model:

```{literalinclude} ../../../../../samples/models/sklearn-iris-minio.yaml
:language: yaml
```

### Google Storage Example

An example for [Google Storage](https://rclone.org/googlecloudstorage/) could use a [service account](https://cloud.google.com/iam/docs/service-accounts) with credentials created with the [gcloud CLI](https://cloud.google.com/sdk/gcloud/reference/iam/service-accounts/keys/create) as follows:

```bash
gcloud iam service-accounts keys create gcloud-application-credentials.json --iam-account [SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com
```

Once the `gcloud-application-credentials.json` has been created you can then include it in the rclone provider credentials.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gcs-bucket
type: Opaque
stringData:
  gcs: |
    type: gcs
    name: gcs
    parameters:
      service_account_credentials: '<gcloud-application-credentials.json>'
```

and then one could use these in Models like:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mymodel
spec:
  storageUri: "gcs://my-bucket/my-path/my-pytorch-model"
  secretName: "gcs-bucket"
  requirements:
  - pytorch
```

## Central Config Map

To allow all models to utilize particular rclone providers one can add the secrets to the agent configMap, e.g.

```{literalinclude} ../../../../../samples/auth/agent.yaml
:language: yaml
```


