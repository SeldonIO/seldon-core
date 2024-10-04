# Ansible Installation

{% hint style="warning" %}
The Ansible installation of a Seldon Core and associated ecosystem is meant for **dev/testing** purposes.
For production use cases follow [Helm installation](https://docs.seldon.io/projects/seldon-core/en/v2/contents/getting-started/kubernetes-installation/helm.html).
{% endhint %}

## Installing Ansible

Provided Ansible playbooks and roles depends on [kubernetes.core](https://github.com/ansible-collections/kubernetes.core)
Ansible Collection for performing `kubectl` and `helm` operations. Check Ansible [documentation] for further information.

To install Ansible and required collections

```bash
pip install ansible openshift kubernetes docker
ansible-galaxy collection install kubernetes.core
```

We have tested provided instructions on Python 3.8 - 3.11 with following version of Python libraries

| Python | Ansible | Docker | Kubernetes |
|--------|---------|--------|------------|
| 3.8    | 6.7.0   | 6.0.1  | 26.1.0     |
| 3.9    | 7.2.0   | 6.0.1  | 26.1.0     |
| 3.10   | 7.2.0   | 6.0.1  | 26.1.0     |
| 3.11   | 7.2.0   | 6.0.1  | 26.1.0     |

and `kubernetes.core` collection in version `2.4.0`.


Once installed you can use the following Playbooks that you will find in
[Ansible](https://github.com/SeldonIO/seldon-core/tree/v2/ansible) folder of Seldon Core V2 repository.

You also need to have installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) CLI.


## Installing Seldon Core v2 using Ansible

### One-liner local kind install (from scratch)

If you simply want to install into a fresh local [kind](https://kind.sigs.k8s.io/)
k8s cluster, the `seldon-all` playbook allows you to do so with a single command:

```bash
ansible-playbook playbooks/seldon-all.yaml
```

This will create a Kind cluster and install ecosystem dependencies (kafka,
prometheus, opentelemetry, jager) as well as all the seldon-specific components.
The seldon components are installed using helm-charts from the current git
checkout (`../k8s/helm-charts/`).

Internally this runs, in order, the following playbooks (described in more detail
in the sections below):
- kind-cluster.yaml
- setup-ecosystem.yaml
- setup-seldon.yaml

You may pass any of the additonal variables which are configurable for those playbooks
to `seldon-all`. See the Customizing Ansible Instalation section for details.

For example:

```bash
ansible-playbook playbooks/seldon-all.yaml -e seldon_mesh_namespace=my-seldon-mesh -e install_prometheus=no -e @playbooks/vars/set-custom-images.yaml
```

Running the playbooks individually, as described in the sections below, will give you
more control over what gets run and when (for example, if you want to install into an
existing k8s cluster).


### Create Kind Cluster

It is recommended to first install Seldon Core v2 inside Kind cluster.
This allow to test and trial the installation in isolated environment that is easy to remove.

```bash
ansible-playbook playbooks/kind-cluster.yaml
```

### Setup Ecosystem

Seldon runs by default in `seldon-mesh` namespace and a Jaeger pod and  and OpenTelemetry
collector are installed in the chosen namespace. Run the following:

```bash
ansible-playbook playbooks/setup-ecosystem.yaml
```

The most common change will be to install in another namespace with:

```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e seldon_mesh_namespace=<mynamespace>
```

### Install Seldon Core v2

Run the following from the `ansible/` folder:

```bash
ansible-playbook playbooks/setup-seldon.yaml
```

If you have changed the namespace you wish to use you will need to run with:

```bash
ansible-playbook playbooks/setup-seldon.yaml -e seldon_mesh_namespace=<mynamespace>
```


## Customizing Ansible Installation

### Ecosystem configuration options

The ecosystem setup can be parametrized by providing extra Ansible variables, e.g. using `-e`
flag to `ansible-playbook` command.

For example run the following from the `ansible/` folder:
```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e full_install=no -e install_kafka=yes
```
will only install Kafka when setting up the ecosystem.

|                         | type   | default                       | comment                                                  |
|-------------------------|--------|-------------------------------|----------------------------------------------------------|
| seldon_mesh_namespace   | string | seldon-mesh                   | namespace to install Seldon Core v2                      |
| seldon_kafka_namespace  | string | seldon-mesh                   | namespace to install Kafka Cluster for Core v2           |
| full_install            | bool   | yes                           | enables full ecosystem installation                      |
| install_kafka           | bool   | `{{ full_install }}`          | installs Strimzi Kafka Operator                          |
| install_prometheus      | bool   | `{{ full_install }}`          | installs Prometheus Operator                             |
| install_grafana         | bool   | `{{ full_install }}`          | installs Grafana Operator                                |
| install_certmanager     | bool   | `{{ full_install }}`          | installs Cert Manager                                    |
| install_jaeger          | bool   | `{{ full_install }}`          | installs Jaeger                                          |
| install_opentelemetry   | bool   | `{{ full_install }}`          | installs OpenTelemetry                                   |
| configure_kafka         | bool   | `{{ install_kafka }}`         | configures Kafka Cluster for Core v2                     |
| configure_prometheus    | bool   | `{{ install_prometheus }}`    | configure Prometheus using Core v2 specific resources    |
| configure_jaeger        | bool   | `{{ install_jaeger }}`        | configure Jaeger using Core v2 specific resources        |
| configure_opentelemetry | bool   | `{{ install_opentelemetry }}` | configure OpenTelemetry using Core v2 specific resources |



### Seldon Core v2 configuration options

|                         | type   | default                       | comment                                                 |
|-------------------------|--------|-------------------------------|---------------------------------------------------------|
| seldon_kafka_namespace  | string | seldon-mesh                   | namespace to install Kafka                              |
| seldon_mesh_namespace   | string | seldon-mesh                   | namespace to install Seldon                             |
| seldon_crds_namespace   | string | default                       | namespace to install Seldon CRDs                        |`
| full_install            | bool   | yes                           | enables full ecosystem installation                     |
| install_crds            | bool   | `{{ full_install }}`          | installs Seldon CRDs                                    |
| install_components      | bool   | `{{ full_install }}`          | install Seldon components                               |
| install_servers         | bool   | `{{ full_install }}`          | install Seldon servers                                  |


#### Custom Seldon images and private registries

By default, the container images used in the install are the ones defined by the helm
charts (referring to images publicly available on dockerhub).

If you need to customize the images (i.e pull from private registry, pull given
tag), create a custom images config file following the example in
`playbooks/vars/set-custom-images.yaml` and run with:

```bash
ansible-playbook playbooks/setup-seldon.yaml -e @<path-to-custom-images-config.yaml>
```

If, instead of pulling images from an external repository you want to build certain components
locally, please read README.dev.md

##### Private registries

When using private registries, access needs to be authenticated (typically, via a
service account key), and the k8s cluster will need to have access to a secret holding
this key to be able to pull images.

The `setup-seldon.yaml` playbook will create the required k8s secrets inside the
cluster if it is provided with an auth file in `dockerconfigjson` format. You provide
the path to this file by cusomizing the `custom_image_pull_secrets.dockerconfigjson`
variable and define the secret name via `custom_image_pull_secrets.name` in the custom
images config file (the one passed to the playbook via `-e @file`).

By default, docker creates the `dockerconfigjson` auth file in `~/.docker/config.json`
after passing the service-account key to `docker login`.

The `docker login` command would look like this (key in json format):

```bash
cat registry-sa-key.json | docker login -u _json_key --password-stdin <registry-url>
```

or, for keys in base64 format:

```bash
cat registry-sa-key | docker login -u _json_key_base64 --password-stdin <registry-url>
```

##### Saving helm-chart customisations

Because the additional custom images config file (starting from the
`playbooks/vars/set-custom-images.yaml` example) overrides values in the helm-charts
available in the repo, there's also a playbook option of saving those overrides as a
separate values file, which could be used if deploying manually via helm.

This is controlled via two variables:

|                                      | type   | default                        | comment                             |
|--------------------------------------|--------|--------------------------------|-------------------------------------|
| save_helm_components_overrides       | bool   | false                          | enable saving helm values overrides |
| save_helm_components_overrides_file  | string | ~/seldon_helm_comp_values.yaml | path/filename for saving overrides  |

You can either pass those within the custom images config file or directly when running the
playbook. For example, for just saving the helm-chart overrides (without installing seldon
components), you would run:

```bash
ansible-playbook playbooks/setup-seldon.yaml -e full_install=no -e save_helm_components_overrides=yes -e @<path-to-custom-images-config.yaml>
```

Please note that when deploying outside ansible via helm using this saved overrides file,
and using private registries, you will have to manually create the service-account key
secret with the same name as the one defined in your custom image config file under
`custom_image_pull_secrets.name`.

## Uninstall

To fully remove the Ansible installation delete the created Kind cluster

```bash
kind delete cluster --name seldon
```

This will stop and delete the `Kind` cluster freeing all of the resources taken by the dev/trial installation.
You may want to also remove cache resources used for the installation with

```bash
rm -rf ~/.cache/seldon/
```

{% hint style="info" %}
If you used Ansible to install Seldon Core v2 and its ecosystem into K8s cluster other than Kind you need to manually remove all the components.
Notes on how to remove Seldon Core v2 Helm installation itself you can find [here](https://docs.seldon.io/projects/seldon-core/en/v2/contents/getting-started/kubernetes-installation/helm.html#uninstall).
{% endhint %}
