# Server

{% hint style="info" %}
The default installation will provide two initial servers: one MLServer and one Triton. You only need to define additional servers for advanced use cases.
{% endhint %}

A Server defines an inference server onto which models will be placed for inference. By default on installation two server StatefulSets will be deployed one MlServer and one Triton. An example Server definition is shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver
spec:
  serverConfig: mlserver
  replicas: 1
```

The main requirement is a reference to a ServerConfig resource in this case `mlserver`.

## Detailed Specs

```go
type ServerSpec struct {
	// Server definition
	ServerConfig string `json:"serverConfig"`
	// The extra capabilities this server will advertise
	// These are added to the capabilities exposed by the referenced ServerConfig
	ExtraCapabilities []string `json:"extraCapabilities,omitempty"`
	// The capabilities this server will advertise
	// This will override any from the referenced ServerConfig
	Capabilities []string `json:"capabilities,omitempty"`
	// Image overrides
	ImageOverrides *ContainerOverrideSpec `json:"imageOverrides,omitempty"`
	// PodSpec overrides
	// Slices such as containers would be appended not overridden
	PodSpec *PodSpec `json:"podSpec,omitempty"`
	// Scaling spec
	ScalingSpec `json:",inline"`
	// +Optional
	// If set then when the referenced ServerConfig changes we will NOT update the Server immediately.
	// Explicit changes to the Server itself will force a reconcile though
	DisableAutoUpdate bool `json:"disableAutoUpdate,omitempty"`
}

type ContainerOverrideSpec struct {
	// The Agent overrides
	Agent *v1.Container `json:"agent,omitempty"`
	// The RClone server overrides
	RClone *v1.Container `json:"rclone,omitempty"`
}

type ServerDefn struct {
	// Server config name to match
	// Required
	Config string `json:"config"`
}
```

## Custom Servers

One can easily utilize a custom image with the existing ServerConfigs. For example, the following defines an MLServer server with a custom image:


```yaml
# samples/servers/custom-mlserver.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-134
spec:
  serverConfig: mlserver
  extraCapabilities:
  - mlserver-1.3.4
  podSpec:
    containers:
    - image: seldonio/mlserver:1.3.4
      name: mlserver
```

This server can then be targeted by a particular model by specifying this server name when creating the model, for example:


```yaml
# samples/models/iris-custom-server.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  server: mlserver-134
```

### Server with PVC

One can also create a Server definition to add a persistent volume to your server. This can be used to allow
models to be loaded directly from the persistent volume.


```yaml
# samples/examples/k8s-pvc/server.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-pvc
spec:
  serverConfig: mlserver
  extraCapabilities:
  - "pvc"
  podSpec:
    volumes:
    - name: models-pvc
      persistentVolumeClaim:
        claimName: ml-models-pvc
    containers:
    - name: rclone
      volumeMounts:
      - name: models-pvc
        mountPath: /var/models
```

The server can be targeted by a model whose artifact is on the persistent volume as shown below.

```yaml
# samples/examples/k8s-pvc/iris.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "/var/models/iris"
  requirements:
  - sklearn
  - pvc
```

A fully worked example for this can be found [here](../../examples/k8s-pvc.md).

An alternative would be to create your own [ServerConfig](./serverconfig.md) for more complex use cases or you
want to standardise the Server definition in one place.
