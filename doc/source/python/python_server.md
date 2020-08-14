# Seldon Python Server

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
