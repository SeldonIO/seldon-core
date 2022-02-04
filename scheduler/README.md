# Seldon Envoy XDS Server


## Development Setup

The scheduler and its associated control-plane components in Seldon Core v2 are intended to run both in and out of Kubernetes.
While it is possible to run all the individual components locally as plain, OS-level processes (see below), the recommended approach is to use Docker Compose to ensure isolation and reproducibility.

### Docker Compose

#### Installation

Ensure that you are using a version of Docker Compose ("Compose") [compatible with the 3.9+ file format](https://docs.docker.com/compose/compose-file/compose-file-v3/#compose-and-docker-compatibility-matrix).

You can install directly by following [the official instructions](https://docs.docker.com/compose/install/).
Alternatively, there are `asdf` plugins for Compose which you can choose from:

```bash
asdf plugin-list-all | grep docker
```

You may also want to install shell completion, if it has not been installed for you.
See the [official documentation](https://docs.docker.com/compose/completion/).

#### Full control plane

There are Makefile targets to start and stop a minimal Seldon Core v2 setup.
You need to choose whether to run with MLServer or Triton as the inference server, but otherwise the setups are almost identical.

To get started:

```bash
make start-all-mlserver
# OR
make start-all-triton
```

and once finished, you can spin down the services with:

```bash
make stop-all-mlserver
# OR
make stop-all-triton
```

#### Examples

It is possible to interact with and even start or stop individual services when using Compose.
A collection of examples are given below.

> **NOTE** You need to specify the project (`-p`) and the appropriate manifest and environment files when interacting directly with services start via the Makefile targets.

Compose allows multiple manifests to be specified, with the latter ones able to [override configuration](https://docs.docker.com/compose/extends/#multiple-compose-files) from the earlier ones.
This greatly reduces duplication between the MLServer and Triton setups at the cost of an extra command-line argument.

<details>
  <summary>Check running services</summary>

  ```bash
  $ docker-compose -f all-base.yaml -f all-mlserver.yaml --env-file env.all -p scv2_mlserver ps
            Name                         Command               State                                            Ports
  -------------------------------------------------------------------------------------------------------------------------------------------------------------
  scv2_mlserver_agent_1       /bin/agent --log-level deb ...   Up      0.0.0.0:8090->8090/tcp,:::8090->8090/tcp, 0.0.0.0:8091->8091/tcp,:::8091->8091/tcp
  scv2_mlserver_envoy_1       /docker-entrypoint.sh /bin ...   Up      10000/tcp, 0.0.0.0:9000->9000/tcp,:::9000->9000/tcp,
                                                                       0.0.0.0:9003->9003/tcp,:::9003->9003/tcp
  scv2_mlserver_rclone_1      rclone rcd --rc-no-auth -- ...   Up      0.0.0.0:5572->5572/tcp,:::5572->5572/tcp
  scv2_mlserver_scheduler_1   /bin/scheduler                   Up      0.0.0.0:9002->9002/tcp,:::9002->9002/tcp, 0.0.0.0:9004->9004/tcp,:::9004->9004/tcp,
                                                                       0.0.0.0:9005->9005/tcp,:::9005->9005/tcp
  scv2_mlserver_server_1      mlserver start /mnt/models       Up      0.0.0.0:8080->8080/tcp,:::8080->8080/tcp, 0.0.0.0:8081->8081/tcp,:::8081->8081/tcp
  ```
</details>

<details>
  <summary>Check logs</summary>

  ```bash
  $ docker-compose -f all-base.yaml -f all-mlserver.yaml --env-file env.all -p scv2_mlserver logs agent | tail
  agent_1      | time="2022-02-04T12:14:03Z" level=info msg="Calling Rclone server: /rc/noop with {\"foo\":\"bar\"}" Source=RCloneClient
  agent_1      | time="2022-02-04T12:14:03Z" level=error msg="Rclone not ready" Name=Client error="Post \"http://0.0.0.0:5572/rc/noop\": dial tcp 0.0.0.0:5572: connect: connection refused" func=waitReady
  ...
  ```
</details>

<details>
  <summary>Restart a service</summary>

  The argument to the `restart` command is the name of the **service** in the manifest, e.g. `server`.

  ```bash
  $ docker-compose -f all-base.yaml -f all-mlserver.yaml --env-file env.all -p scv2_mlserver restart server
  Restarting scv2_mlserver_server_1 ... done
  ```
</details>

<details>
  <summary>Start a single service</summary>

  Note that this example does not specify a `project`, so Compose defaults to the parent directory's name.
  This example also does not specify `-f all-mlserver.yaml` as the Rclone configuration is defined in `all-base.yaml` and is completely independent of the inference server in use.

  We can override values from `env.all` to specify the repository and tag for the Rclone image.

  ```bash
  $ RCLONE_IMAGE_AND_TAG=seldonio/seldon-reclone:latest docker-compose -f all-base.yaml --env-file env.all run -d rclone
  Building rclone
  Step 1/3 : FROM rclone/rclone:1.56.2
  1.56.2: Pulling from rclone/rclone
  a0d0a0d46f8b: Already exists
  ...
  Creating scheduler_rclone_run ... done
  scheduler_rclone_run_4379d7918894
  ```
</details>

<details>
  <summary>Inspect environment variables in a container</summary>

  It can be useful to see what environment variables have been set in a container, as these do not show up in the command column of `ps` commands.
  The below shows a way of inspecting these values for a container called `scv2_mlserver_agent_1`:

  ```bash
  $ docker inspect -f '{{ range $i, $v := .Config.Env }}{{ $v }}{{ println }}{{ end }}' scv2_mlserver_agent_1
  SELDON_OVERCOMMIT=false
  SELDON_SERVER_HTTP_PORT=8090
  SELDON_SERVER_GRPC_PORT=8091
  SELDON_DEBUG_GRPC_PORT=7777
  SELDON_SCHEDULER_HOST=0.0.0.0
  SELDON_SCHEDULER_PORT=9005
  MEMORY_REQUEST=1000000
  SELDON_SERVER_TYPE=mlserver
  SELDON_SERVER_CAPABILITIES=sklearn,xgboost
  PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
  GOLANG_VERSION=1.17.6
  GOPATH=/go
  ```
</details>

<details>
  <summary>Check specific details for running containers</summary>

  The default view of `docker-compose ps` is limited, in that it does not have all the same configuration values as the equivalent `docker` command.
  It can also be very dense for a split-screen or smaller screen view.
  The below provides an example on customising the display:

  ```bash
  $ docker ps --format 'table {{ .Image }}\t{{ .Names }}\t{{ .Status }}\t{{ .Command }}' --no-trunc
  IMAGE                                NAMES                    STATUS          COMMAND
  seldonio/seldon-envoy-local:latest   scv2_mlserver_envoy_1    Up 2 hours      "/docker-entrypoint.sh /bin/sh -c '/usr/local/bin/envoy -c /etc/envoy.yaml'"
  seldonio/seldon-rclone:latest        scv2_mlserver_rclone_1   Up 2 hours      "rclone rcd --rc-no-auth --config=/rclone/rclone.conf --rc-addr=0.0.0.0:5572 --verbose"
  seldonio/mlserver:1.0.0.rc1          scv2_mlserver_server_1   Up 10 minutes   "mlserver start /mnt/models"
  registry:2                           kind-registry            Up 7 hours      "/entrypoint.sh /etc/docker/registry/config.yml"
  ```
</details>


### gRPC compile

Install [protoc](https://github.com/protocolbuffers/protobuf/releases).

```
protoc --version
libprotoc 3.18.0
```

Intall [go grpc plugins](https://grpc.io/docs/languages/go/quickstart/)

```
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
```

### Local Envoy Install

Install or [download](https://archive.tetratelabs.io/envoy/envoy-versions.json) envoy.

Tested with:

```
envoy --version

envoy  version: a2a1e3eed4214a38608ec223859fcfa8fb679b14/1.19.1/Clean/RELEASE/BoringSSL
```

### MlServer

```
pip install mlserver
```


## Local Test

```
make build
```

Follow steps in [local test notebook](./notebooks/scheduler-local-test.ipynb)


## K8S Test

```
make kind-image-install-all
```

Follow steps in [k8s test notebook](./notebooks/scheduler-k8s-test.ipynb)


## Docs

[development docs](./docs/README.md)

