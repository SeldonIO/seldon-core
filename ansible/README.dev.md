# Creating and using development image configs

This documents and gives examples on how you can write custom configuration files that allow
you to use local dev builds with your ansible kind installs, as well as customize installation
components.

Locally built images use our usual docker image building process, and the resulting images are
then copied into the kind cluster and correctly referenced from within the helm charts, without
requiring additional intervention.

```{warning}
Please install Ansible as described in README.md before following the instructions in this file
```

In all the cases described below, save your configuration into a file (e.g.
`~/seldon-configs/my-dev-config.yaml`) then pass this file as an extra configuration variables
file when running ansible playbooks. Please note the `@` before the file path, which is required:

```bash
ansible-playbook playbooks/seldon-all.yaml -e @<path-to-custom-images-config.yaml>
```

By creating multiple such config files under a separate directory, you can easily switch between
sets of dev configurations when deploying.


## Minimal dev image config

Say you want to use all the default images from the normal helm install, with the exception of
TWO components, which you've modified and you would like to build and test locally.

In this case, having a configuration file like the example below is sufficient:

```yaml

seldon_dev:
  dockerhub_user: seldondev

custom_image_config:
  - dev_img_set: true
    components:
      - dataflow
      - pipelinegateway
    image:
      tag: "pipeline-test"

```
When passing the configuration file to the `seldon-all.yaml` or the `setup-seldon.yaml` ansible
playbooks:

* Two new docker images are built:
 - `seldondev/dataflow:pipeline-test`
 - `seldondev/pipelinegateway:pipeline-test`

* Those images get copied into the kind install
* The installed helm-charts end up pointing to those kind-local docker images
* All other components are configured to their defaults (as defined in the helm-charts in your
  worktree)

The configurable components (which you can specify in the `components` list) are:
    - scheduler
    - modelgateway
    - pipelinegateway
    - dataflow
    - controller
    - hodometer
    - envoy
    - agent
    - rclone
    - grafana
    - mlserver

It's highly recommended you configure a custom `dockerhub_user` as above. If you don't, the
default will be `seldonio`, and depending on what tag you configure you might end up shadowing one
of the docker images that were pulled locally from the upstream docker registry.

Running the same playbook a second time will by default rebuild & copy the images into kind.
If you would like to skip the time-consuming build step for images that already exist (they
match the name of previously built images or of images that already exist locally in docker),
then you may set the `seldon_dev.build_skip_existing_images` variable to `true`:

```yaml

seldon_dev:
  dockerhub_user: seldondev
  build_skip_existing_images: true

custom_image_config:
  - dev_img_set: true
    components:
      - dataflow
      - pipelinegateway
    image:
      tag: "pipeline-test"
```

## Mixed-tag configurations

Sometimes, you might want to configure sets of components using different tags, perhaps using dev
images that were previously built for some components, and building some new ones for others.
This is possible:

```yaml

seldon_dev:
  dockerhub_user: seldondev
  build_skip_existing_images: true

custom_image_config:
  - dev_img_set: true
    components:
      - dataflow
    image:
      tag: "pipeline-test"
  - dev_img_set: true
    components:
      - pipelinegateway
    image:
      tag: "pipeline-test-old"
```

## Mixing dev images and non-dev custom images

If you want to use some locally-built components and some external components from dockerhub but
not the tags included by default in your helm charts, you can also configure that:

```yaml

seldon_dev:
  dockerhub_user: seldondev
  build_skip_existing_images: true

custom_image_config:
  - dev_img_set: true
    components:
      - dataflow
      - pipelinegateway
    image:
      tag: "pipeline-test"

  - components:
      - modelgateway
      - scheduler
      - controller
    image:
      tag: "2.7.0"
```

Please note we're no longer setting `dev_img_set` for the second element of the
`custom_image_config` list, which indicates we're not going to build those images locally but
try to fetch them from dockerhub.


## Using non-dev images from private registries

If you would like to fetch some of the components from private registries:

```yaml

seldon_dev:
  dockerhub_user: seldondev
  build_skip_existing_images: true

custom_image_config:
  - dev_img_set: true
    components:
      - dataflow
      - pipelinegateway
    image:
      tag: "pipeline-test"

  - components:
      - mlserver
    repository_img_prefix: ""
    image:
      registry: [...pkg.dev/dev-seldon-registry/mlserver]
      tag: 1.4.0.rc9

custom_image_pull_secrets:
  name: private-registry
  dockerconfigjson: ~/.docker/config.json
```

Note the `repository_img_prefix` set to "" when using the private registry: depending on the
configuration of your private registry, you will not want the default `seldonio` image prefix.

For private repository definitions, each component will have its own name, depending
on the registry configuration. You may configure custom names by defining a top-level
`custom_image_repository` variable:

```yaml
# ... other component configs

custom_image_repository:
  hodometer: "seldon-hodometer"
  modelgateway: "seldon-modelgateway"
  pipelinegateway: "seldon-pipelinegateway"
  dataflow: "seldon-dataflow-engine"
  controller: "seldonv2-controller"
  envoy: "seldon-envoy"
  scheduler: "seldon-scheduler"
  rclone: "seldon-rclone"
  agent: "seldon-agent"
  mlserver: "mlserver"
```

You will also need to define the custom image pull secret to be created or updated in kind for
access to the private registry.

This assumes you already have the service account created and you have fetched
its key in JSON format. Authenticate the docker CLI with this key, i.e:

> cat access-seldon-registry.json | docker login -u _json_key --password-stdin [....registry.pkg.dev]

then configure the dockerconfigjson property in `custom_image_pull_secrets` with the resulting
docker config.json (the default path is the one set above, `~/.docker/config.json`)


## Configuring secrets to be installed into kind

Sometimes, you might want your kind install to contain secrets. It's possible to automatically
create those secrets via ansible by defining them in the `custom_secrets` top-level variable:

```yaml
# ... other component configs

custom_secrets:
    - name: my-secret
      template: "~/local/path/to/my-secret.yaml.j2"
      namespaces:
        - "seldon-mesh"
        - "default"
    - name: top-secret
      template: "~/local/path/to/top-secret.yaml.j2"
```

The template is a normal k8s secret yaml (as you would write to create a secret via kubectl)
but without keys like the name or the namespace, which will be added by ansible. For example:

```yaml
apiVersion: v1
kind: Secret
type: Opaque
stringData:
  method: OIDC
  client_id: a-secret-client-id
```

## Mounting local (host) path into the rclone container of a Server pod

For this, you first need to enable local mounts into kind (configuring the `kind_local_mount`,
`kind_host_path` and `kind_container_path` variables). Then, you can use `custom_servers_values`
to set-up the volume & volumeMounts for the inference server pod:

```yaml
helm_force_install: true

kind_local_mount: true
kind_host_path: "<local-path>"
kind_container_path: "/host-models"

seldon_dev:
  install_kind_images: true

custom_image_config:
  - components:
      - hodometer
      - modelgateway
      - pipelinegateway
      - dataflow
      - controller
      - scheduler
      - envoy
      - rclone
      - agent
    image:
      tag: 2.8.3
  - components:
      - mlserver
    image:
      tag: 1.6.1

custom_servers_values:
  mlserver:
    replicas: 1
    podSpec:
      containers:
        - name: rclone
          volumeMounts:
            - name: host-models
              mountPath: "/mnt/local-models"
      volumes:
        - name: host-models
          hostPath:
            path: "{{ kind_container_path }}"
```

You can now load a model into kind with `storageUri` pointing to `/mnt/local-models/...`:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
    name: iris
    namespace: seldon-mesh
spec:
    storageUri: "/mnt/local-models/iris"
    requirements:
    - sklearn
    memory: 100Ki
```

## Complex component configurations

Your config file may overwrite any of the playbook variables described in `README.md`, and use
this to create define complex kinds of configurations. Here, I'll show how one may configure a
setup where instead of strimzi kafka we use Confluent Cloud:

```yaml
install_kafka: false
configure_kafka: false

seldon_dev:
  dockerhub_user: seldondev
  build_skip_existing_images: true

random_suffix8: "{{ lookup('password', '/dev/null chars=ascii_lowercase,digits length=8') }}"

custom_image_config:
  - dev_img_set: true
    components:
      - dataflow
    image:
      tag: "dev.{{ random_suffix8 }}"

  - components:
      - controller
      - hodometer
      - agent
      - scheduler
      - modelgateway
      - pipelinegateway
    image:
      tag: "2.8.1-rc2"

  - components:
      - mlserver
    image:
      tag: 1.6.1

custom_components_values:
  kafka:
    bootstrap: broker.europe-west2.gcp.confluent.cloud:9092
    consumer:
      messageMaxBytes: 8388608
    producer:
      messageMaxBytes: 8388608
    topics:
      replicationFactor: 2
      numPartitions: 4
  security:
    kafka:
      protocol: SASL_SSL
      sasl:
        mechanism: OAUTHBEARER
        client:
            secret: kafka-oauth
      ssl:
        client:
          secret:
          brokerValidationSecret:
          endpointIdentificationAlgorithm:

custom_runtime_values:
  config:
    agentConfig:
      rclone:
        configSecrets:
          - gcs-buckets

custom_servers_values:
  mlserver:
    replicas: 1
  triton:
    replicas: 0

custom_secrets:
  - name: "{{ custom_components_values.security.kafka.sasl.client.secret }}"
    template: "path/to/secrets/confluent.yaml.j2"
  - name: "{{ custom_runtime_values.config.agentConfig.rclone.configSecrets[0] }}"
    template: "path/to/secrets/gcs-bucket.yaml.j2"

```

Here, we're not installing or configuring the usual strimzi kafka, we're always generating a new
local build image for the dataflow component (due to it getting a random tag every time), we're
pulling specific tags for various sets of components, configuring kafka to point to a
confluent-cloud broker, set up consumer/producer properties as well as broker options. The kafka
config references a k8s secret name, so wer're also creating that secret in kind from a local
template file.


## Detailed description of available options and defaults

### Override any of the seldon ansible playbook variables

Example:
```yaml
install_kafka: false
configure_kafka: false
```


### seldon_dev

The `seldon_dev` dictionary holds dev options relating to
- image building & deployment into a kind k8s custer
- saving the current customisation as a helm values file for reuse outside ansible

Not providing `seldon_dev` at all is the same as leaving all keys to their default values.

The following keys may be configured:

    - dockerhub_user:    [optional, default: seldonio]
                        The name of the registry user to be used when building images

    - build_skip_existing_images: [optional, default: false]
                        Whether to skip builing a docker image from local sources if one with
                        the same tag already exists as a local docker image.

                        Only applies to components with `dev_img_set: true` in
                        `custom_image_config` (documented further down)

    - install_kind_images: [optional, default: true]
                        Whether to copy the configured dev images into the local kind
                        cluster

                        Only applies to components with `dev_img_set: true` in
                        `custom_image_config` (documented further down)

    - save_helm_components_overrides: [optional, default: false]
                                    Whether to save a helm values file with images
                                    pre-configured to the config described in
                                    `custom_image_config` below. Useful for using outside
                                    ansible

    - save_helm_components_overrides_file: [optional, default: "${HOME}/seldon_helm_comp_values.yaml"]
                                        Path & filename for saving helm value overrides

Example:
```yaml
seldon_dev:
  dockerhub_user: seldondev
  build_skip_existing_images: false
  install_kind_images: true
  save_helm_components_overrides: false
```

### custom_image_config

By configuring `custom_image_config`, you customise image properties for sets of seldon
components. For example, fetching images with particular tags, configuring private registries
or referring to local dev sources.

Any components not specified will be deployed with their default setup.

This is a list, where each item is a dictionary with the following keys:
- dev_img_set : [optional, default: false]
                Whether the components listed in `components` are to be considered local dev
                images that should be built & deployed (when true) or pre-built images to be
                fetched externally (when false)

                If `seldon_dev.build_skip_existing_images` is true, `dev_img_set is true` and
                an image with the given tag is not present amongst the local docker images for
                a component in the `components` list, that image will be built. If all images
                exist, those will be used.

                Set `seldon_dev.build_skip_existing_images` to false to force images to be
                rebuilt even if local docker images with the same tag already exist.

- components: [required, default: []]
            List of components with a common image customisation. Possible list items are:
            - hodometer
            - modelgateway
            - pipelinegateway
            - dataflow
            - controller
            - envoy
            - scheduler
            - mlserver
            - triton (supported with `dev_img_set: false` only)
            - rclone
            - agent

- image: [optional] the keys to overwrite in the existing "image" helm-chart values for each
        component in the `components` list

- repository_img_prefix: [optional, default: `seldon_dev.dockerhub_user`]
                        The repository prefix that should be applied to the docker image
                        name. Typically the dockerhub user.

Example:
```yaml
custom_image_config:

  - dev_img_set: true
    components:
      - dataflow
    image:
      tag: "dev.seldon.df-test"   # when dev_img_set is true, an image with this tag is built
                                  # locally as `[seldon_dev.dockerhub_user]/[component-name]:
                                  # dev.seldon.df-test` for each of the listed components

  - dev_img_set: true
    components:
      - modelgateway
      - pipelinegateway
      - scheduler
      - controller
    image:
      tag: "dev.seldon.core-test" # different set of components can be built locally with
                                  # a different tag
  - components:
      - hodometer
      - envoy
    repository_img_prefix: "seldonio"  # use upstream images for those components, namely
                                       # seldonio/[component-name]:latest
  - components:
      - agent
    image:
      tag: "2.7.0"
    repository_img_prefix: "seldonio"  # use upstream image for this component, with custom tag
                                       # seldonio/[component-name]:2.7.0
  - components:
      - mlserver
    repository_img_prefix: ""
    image:
      registry: private-registry/mlserver  # image from private-registry. also configure
                                           # `custom_image_pull_secrets` to ensure that k8s has
                                           # access to the right secret for pulling this image
      tag: 1.4.0.rc9
```

### custom_components_values

`custom_components_values` provides custom config for the seldon-core-v2-setup helm chart
see [core_v2_src]/k8s/helm-charts/seldon-core-v2-setup/values.yaml.template
for possible configuration options

Image-related options (key/value pairs defined under the `image` key) also configured in
`custom_image_config` above take precedence to the ones defined here.

The example below sets a confluent cloud cluster as the kafka deployment, together
with OAUTH authentication. Typically this will also require setting a secret containing
authentication details (see `custom_secrets` below):

Example:
```yaml
 custom_components_values:
   kafka:
     bootstrap: pkc-l6wr6.europe-west2.gcp.confluent.cloud:9092
     topics:
       replicationFactor: 3
       numPartitions: 4
     consumer:
       messageMaxBytes: 8388608
     producer:
       messageMaxBytes: 8388608
   security:
     kafka:
       protocol: SASL_SSL
       sasl:
         mechanism: OAUTHBEARER
         client:
             secret: kafka-oauth
       ssl:
         client:
           secret:
           brokerValidationSecret:
           endpointIdentificationAlgorithm: https
```


### custom_runtime_values

`custom_runtime_values` provides custom config for the seldon-core-v2-runtime helm chart
see [core_v2_src]/k8s/helm-charts/seldon-core-v2-runtime/values.yaml for possible
configuration options

The example below sets a preloaded secret to be added to the `seldon-agent` ConfigMap, so
that all the inference servers managed by the runtime have access to a GCS bucket.
Typically this will also require setting a secret containing service account details (see
`custom_secrets` below; also refer to the [documentation](https://docs.seldon.io/projects/seldon-core/en/v2/contents/kubernetes/storage-secrets/index.html#preloaded-secrets]):

Example:
```yaml
custom_runtime_values:
  config:
    agentConfig:
      rclone:
        configSecrets:
          - gcs-buckets
```


### custom_servers_values

`custom_servers_values` provides custom config for the seldon-core-v2-servers helm chart
see [core_v2_src]/k8s/helm-charts/seldon-core-v2-servers/values.yaml for possible
configuration options

The example below sets the initial number of mlserver and triton inference server replicas:

Example:
```yaml
custom_servers_values:
  mlserver:
    replicas: 1
  triton:
    replicas: 0
```


### custom_image_repository

For private repository definitions, each component will have its own name, depending
on the registry configuration. You may configure custom names in `custom_image_repository`

Set the core component to registry name mapping here; The defaults are the ones listed below
for reference.

Example:
```yaml

 custom_image_repository:
   hodometer: "seldon-hodometer"
   modelgateway: "seldon-modelgateway"
   pipelinegateway: "seldon-pipelinegateway"
   dataflow: "seldon-dataflow-engine"
   controller: "seldonv2-controller"
   envoy: "seldon-envoy"
   scheduler: "seldon-scheduler"
   rclone: "seldon-rclone"
   agent: "seldon-agent"
   mlserver: "mlserver"
```


### custom_image_pull_secrets

Defines the custom pull secrets to be created or updated in k8s for access to private
registries.

This assumes you already have the service account created and you have fetched
its key in JSON format. Authenticate the docker CLI with this key, i.e:

> cat access-seldon-registry.json | docker login -u _json_key --password-stdin [private-reg]

then configure the `dockerconfigjson` property with the resulting docker config.json

Example:
```yaml
custom_image_pull_secrets:
  name: private-registry
  dockerconfigjson: ~/.docker/config.json
```

### custom_secrets

`custom_secrets` defines a list of secrets that should be loaded into various namespaces
of the deployment.

Each element of the list is a dictionary containing the following items:

- name: [mandatory]
        the name of the secret
- template: [mandatory]
            path to a file containing the secret template (yaml, no need to specify name
            and namespace)
- namespaces: [optional, default:[{{ seldon_mesh_namespace }}]]
                list of namespaces to which this secret should be added to

An example below assuming you want to use the same secret name as one defined somewhere
in the custom_components_values variable :

Example:
```yaml
custom_secrets:
  - name: "{{ custom_components_values.security.kafka.sasl.client.secret }}"
    template: "[path-to-secrets-template.yaml]"
    namespaces:
        - "my-namespace-1"
        - "my-namespace-2"
```
