# Seldon Python Server Configuration

 * [Workers](#workers)
 * [Threads](#threads)
 * [Flask Development Server](#development-server)
 * [Server Configuration](#configuration)


To serve your component, Seldon's Python wrapper will use
[Gunicorn](https://gunicorn.org/) under the hood by default.
Gunicorn is a high-performing HTTP server for Unix which allows you to easily
scale your model across multiple worker processes and threads.

.. Note::
  Gunicorn will only handle the horizontal scaling of your model **within the
  same pod and container**.
  To learn more about how to scale your model across multiple pod replicas see
  the :doc:`../graph/scaling` section of the docs.

## Workers

By default, Seldon will only use a **single worker process**.
However, it's possible to increase this number through the `GUNICORN_WORKERS`
environment variable.
This variable can be controlled directly through the `SeldonDeployment` CRD.

For example, to run your model under 4 workers, you could do:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: gunicorn
spec:
  name: worker
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          name: classifier
          env:
          - name: GUNICORN_WORKERS
            value: '4'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```

## Threads

By default, Seldon will process your model's incoming requests using a pool of
**10 threads per worker process**.
You can increase this number through the `GUNICORN_THREADS` environment
variable.
This variable can be controlled directly through the `SeldonDeployment` CRD.

For example, to run your model with 5 threads per worker, you could do:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: gunicorn
spec:
  name: worker
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          name: classifier
          env:
          - name: GUNICORN_THREADS
            value: '5'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```

### Disable multithreading

In some cases, you may want to completely disable multithreading.
To serve your model within a single thread, set the environment variable
`FLASK_SINGLE_THREADED` to 1.
This is not the most optimal setup for most models, but can be useful when your
model cannot be made thread-safe like many GPU-based models that deadlock when
accessed from multiple threads.


```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: flaskexample
spec:
  name: worker
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          name: classifier
          env:
          - name: FLASK_SINGLE_THREADED
            value: '1'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```

## Development server

While Gunicorn is recommended for production workloads, it's also possible to
use Flask's built-in development server.
To enable the development server, you can set the `SELDON_DEBUG` variable to
`1`.

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: flask-development-server
spec:
  name: worker
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          name: classifier
          env:
          - name: SELDON_DEBUG
            value: '1'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```


## Configuration

Python Server can be configured using environmental variables or command line flags.

| CLI Flags               | Environment Variable                   | Default       | Notes                                                                                                                                                              |
|-------------------------|----------------------------------------|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `interface_name`        | N/A                                    | N/A           | First positional argument. Required. If contains `.` first part is interpreted as module name.                                                                     |
| `api_type`              | N/A                                    | `REST`        | Name of interface to use. Can be `REST` or `GRPC`.                                                                                                                 |
| `--port`                | `PREDICTIVE_UNIT_SERVICE_PORT`         | `5000`        | Port of Seldon service. In k8s this is controlled by Seldon Core Operator.                                                                                         |
| `--metrics-port`        | `PREDICTIVE_UNIT_METRICS_SERVICE_PORT` | `6000`        | Metrics port of Seldon service. In k8s this is controlled by Seldon Core Operator.                                                                                 |
| `--service-type`        | N/A                                    | `MODEL`       | Service type of model. Can be `MODEL`, `ROUTER`, `TRANSFORMER`, `COMBINER` or `OUTLIER_DETECTOR`.                                                                  |
| `--parameters`          | N/A                                    | `[]`          | List of parameters to be passed to Model class.                                                                                                                    |
| `--log-level`           | `LOG_LEVEL_ENV`                        | `INFO`        | Python log level. Can be `DEBUG`, `INFO`, `WARNING` or `ERROR`.                                                                                                    |
| `--debug`               | `SELDON_DEBUG`                         | `false`       | Enable debug mode that enables `flask` development server and sets logging to `DEBUG`. Values `1`, `true` or `t` (case insensitive) will be interpreted as `True`. |
| `--tracing`             | `TRACING`                              | `0`           | Enable tracing. Can be `0` or `1`.                                                                                                                                 |
| `--workers`             | `GUNICORN_WORKERS`                     | `1`           | Number of Gunicorn workers for handling requests.                                                                                                                  |
| `--threads`             | `GUNICORN_THREADS`                     | `10`          | Number of threads to run per Gunicorn worker.                                                                                                                      |
| `--max-requests`        | `GUNICORN_MAX_REQUESTS`                | `0`           | Maximum number of requests gunicorn worker will process before restarting.                                                                                         |
| `--max-requests-jitter` | `GUNICORN_MAX_REQUESTS_JITTER`         | `0`           | Maximum random jitter to add to max-requests.                                                                                                                      |
| `--pidfile`             | N/A                                    | None          | A file path to use for the Gunicorn PID file.                                                                                                                      |
| `--single-threaded`     | `FLASK_SINGLE_THREADED`                | `0`           | Force the Flask app to run single-threaded. Also applies to Gunicorn. Can be `0` or `1`.                                                                           |
| N/A                     | `FILTER_METRICS_ACCESS_LOGS`           | `not debug`   | Filter out logs related to Prometheus accessing the metrics port. By default enabled in production and disabled in debug mode.                                     |
| N/A                     | `PREDICTIVE_UNIT_METRICS_ENDPOINT`     | `/metrics`    | Endpoint name for Prometheus metrics. In k8s deployment default is `/prometheus`.                                                                                  |
| N/A                     | `PAYLOAD_PASSTHROUGH`                  | `false`       | Skip decoding of payloads.                                                                                                                                         |
