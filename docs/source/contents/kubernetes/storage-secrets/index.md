# Storage Secrets

Inference artifacts referenced by Models can be stored in any of the storage backends supported by [Rclone](https://rclone.org/).
This includes local filesystems, AWS S3, and Google Cloud Storage (GCS), among others.
Configuration is provided out-of-the-box for public GCS buckets, which enables the use of Seldon-provided models like in the below example:

```{literalinclude} ../../../../../samples/models/sklearn-iris-gs.yaml 
:language: yaml
```

This configuration is provided by the Kubernetes Secret `seldon-rclone-gs-public`.
It is made available to Servers as a [preloaded secret](#preloaded-secrets).
You can define and use your own storage configurations in exactly the same way.

## Configuration Format

To define a new storage configuration, you need the following details:
* Remote name
* Remote type
* Provider parameters

A _remote_ is what Rclone calls a storage location.
The _type_ defines what protocol Rclone should use to talk to this remote.
A _provider_ is a particular implementation for that storage type.
Some storage types have multiple providers, such as `s3` having AWS S3 itself, MinIO, Ceph, and so on.

The remote **name** is your choice.
The prefix you use for models in `spec.storageUri` must be the same as this remote name.

The remote type is one of the values [supported by Rclone](https://rclone.org/docs/).
For example, for AWS S3 it is `s3` and for Dropbox it is `dropbox`.

The provider parameters depend entirely on the remote _type_ and the specific _provider_ you are using.
Please check the Rclone documentation for the appropriate provider.
Note that Rclone docs for storage types call the parameters _properties_ and provide both _config_ and _env var_ formats--you need to use the _config_ format.
For example, the GCS parameter `--gcs-client-id` described [here](https://rclone.org/googlecloudstorage/#gcs-client-id) should be used as `client_id`.

For reference, this format is described in the [Rclone documentation](https://rclone.org/rc/#config-create).
Note that we do not support the use of `opts` discussed in that section.

## Kubernetes Secrets

Kubernetes Secrets are used to store Rclone configurations, or _storage secrets_, for use by Servers.
Each Secret should contain **exactly one** Rclone configuration.

A Server can use storage secrets in one of two ways:
* It can dynamically load a secret specified by a Model in its `.spec.secretName`
* It can use global configurations made available via [preloaded secrets](#preloaded-secrets)

The name of a Secret is entirely your choice, as is the name of the data key in that Secret.
All that matters is that there is a single data key and that its value is in the format described above.

```{note}
It is possible to use preloaded secrets for some Models and dynamically loaded secrets for others.
```

### Preloaded Secrets

Rather than Models always having to specify which secret to use, a Server can load storage secrets ahead of time.
These can then be reused across many Models.

When using a preloaded secret, the Model definition should leave `.spec.secretName` empty.
The protocol prefix in `.spec.storageUri` still needs to match the remote name specified by a storage secret.

The secrets to preload are named in a centralised ConfigMap called `seldon-agent`.
This ConfigMap applies to _all_ Servers managed by the same SeldonRuntime.
By default this ConfigMap only includes `seldon-rclone-gs-public`, but can be extended with your own secrets as shown below:

```{literalinclude} ../../../../../samples/auth/agent.yaml
:language: yaml
```

The easiest way to change this is to update your SeldonRuntime.
* If your SeldonRuntime is configured using the `seldon-core-v2-runtime` Helm chart, the corresponding value is `config.agentConfig.rclone.configSecrets`.
  This can be used as shown below:
  ```yaml
  config:
    agentConfig:
      rclone:
        configSecrets:
          - my-s3
          - custom-gcs
          - minio-in-cluster
  ```
* Otherwise, if your SeldonRuntime is configured directly, you can add secrets by setting `.spec.config.agentConfig.rclone.config_secrets`.
  This can be used as follows:
  ```yaml
  apiVersion: mlops.seldon.io/v1alpha1
  kind: SeldonRuntime
  metadata:
    name: seldon
  spec:
    seldonConfig: default
    config:
      agentConfig:
        rclone:
          config_secrets:
            - my-s3
            - custom-gcs
            - minio-in-cluster
    ...
  ```

## Examples

`````{tabs}

````{group-tab} S3 MinIO

Assuming you have installed MinIO in the `minio-system` namespace, a corresponding secret could be:

```{literalinclude} ../../../../../samples/auth/minio-secret.yaml
:language: yaml
```

You can then reference this in a Model with `.spec.secretName`:

```{literalinclude} ../../../../../samples/models/sklearn-iris-minio.yaml
:language: yaml
```
````

````{group-tab} Google Cloud Storage

[GCS](https://rclone.org/googlecloudstorage/) can use [service accounts](https://cloud.google.com/iam/docs/service-accounts) for access.
You can generate the credentials for a service account using the [gcloud CLI](https://cloud.google.com/sdk/gcloud/reference/iam/service-accounts/keys/create):

```bash
gcloud iam service-accounts keys create \
  gcloud-application-credentials.json \
  --iam-account [SERVICE-ACCOUNT--NAME]@[PROJECT-ID].iam.gserviceaccount.com
```

The contents of `gcloud-application-credentials.json` can be put into a secret:

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

You can then reference this in a Model with `.spec.secretName`:

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
````
`````
