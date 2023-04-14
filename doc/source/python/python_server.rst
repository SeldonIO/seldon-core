Seldon Python Server Configuration
==================================

To serve your component, Seldon's Python wrapper will use
`Gunicorn <https://gunicorn.org/>`__ under the hood by default. Gunicorn
is a high-performing HTTP server for Unix which allows you to easily
scale your model across multiple worker processes and threads.

.. Note:: 
  Gunicorn will only handle the horizontal scaling of your model
  **within the same pod and container**. To learn more about how to scale
  your model across multiple pod replicas see
  `this section <../graph/scaling>`_ of the docs.

Workers
-------

By default, Seldon will only use a **single worker process**. However,
it's possible to increase this number through the ``GUNICORN_WORKERS``
env var for REST and the ``GRPC_WORKERS`` env var for GRPC. 
This variable can be controlled directly through the ``SeldonDeployment`` CRD.

For example, to run your model under 8 processes (4 RESt and 4 GRPC), you could do:

.. code:: yaml
    :emphasize-lines: 14-17

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
              - name: GRPC_WORKERS
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


Running only REST server by disabling GRPC server
-------------------------------------------------------

By default the Seldon models run a REST and GRPC server with a single process each.
If the machine learning model is loaded in each process, this can result in a large 
overhead in cases where the model artifacts are very large as there would be an instance
of the model loaded for each worker. For this case, it is possible to disable the GRPC
server by setting ``GRPC_WORKERS`` to ``0``, which would end up not starting a GRPC server.
It is important to note that the GRPC endpoint will still be available in the service 
orchestrator so GRPC requests would no longer work. An example of this would be as follows:

.. code:: yaml
    :emphasize-lines: 14-15

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
              - name: GRPC_WORKERS
                value: '0'
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


Threads
-------

By default, Seldon will process your model's incoming requests using 
**1 thread per worker process**. You can increase this number
through the ``GUNICORN_THREADS`` environment variable. This variable can
be controlled directly through the ``SeldonDeployment`` CRD.

For example, to run your model with 5 threads per worker, you could do:

.. code:: yaml
    :emphasize-lines: 14-15


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

Disable multithreading
~~~~~~~~~~~~~~~~~~~~~~

In some cases, you may want to completely disable multithreading. To
serve your model within a single thread, set the environment variable
``FLASK_SINGLE_THREADED`` to 1. This is not the most optimal setup for
most models, but can be useful when your model cannot be made
thread-safe like many GPU-based models that deadlock when accessed from
multiple threads.

.. code:: yaml
    :emphasize-lines: 14-15

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

Development server
------------------

While Gunicorn is recommended for production workloads, it's also
possible to use Flask's built-in development server. To enable the
development server, you can set the ``SELDON_DEBUG`` variable to ``1``.

.. code:: yaml
    :emphasize-lines: 14-15

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

Configuration
-------------

Python Server can be configured using environmental variables or command
line flags.

+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| CLI Flags                   | Environment Variable                       | Default         | Notes                                                                                                                                                                            |
+=============================+============================================+=================+==================================================================================================================================================================================+
| ``interface_name``          | N/A                                        | N/A             | First positional argument. Required. If contains ``.`` first part is interpreted as module name.                                                                                 |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--http-port``             | ``PREDICTIVE_UNIT_HTTP_SERVICE_PORT``      | ``9000``        | Http port of Seldon service. In k8s this is controlled by Seldon Core Operator.                                                                                                  |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--grpc-port``             | ``PREDICTIVE_UNIT_GRPC_SERVICE_PORT``      | ``5000``        | Grpc port of Seldon service. In k8s this is controlled by Seldon Core Operator.                                                                                                  |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--metrics-port``          | ``PREDICTIVE_UNIT_METRICS_SERVICE_PORT``   | ``6000``        | Metrics port of Seldon service. In k8s this is controlled by Seldon Core Operator.                                                                                               |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--service-type``          | N/A                                        | ``MODEL``       | Service type of model. Can be ``MODEL``, ``ROUTER``, ``TRANSFORMER``, ``COMBINER`` or ``OUTLIER_DETECTOR``.                                                                      |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--parameters``            | N/A                                        | ``[]``          | List of parameters to be passed to Model class.                                                                                                                                  |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--log-level``             | ``LOG_LEVEL_ENV``                          | ``INFO``        | Python log level. Can be ``DEBUG``, ``INFO``, ``WARNING`` or ``ERROR``.                                                                                                          |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--debug``                 | ``SELDON_DEBUG``                           | ``false``       | Enable debug mode that enables ``flask`` development server and sets logging to ``DEBUG``. Values ``1``, ``true`` or ``t`` (case insensitive) will be interpreted as ``True``.   |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--tracing``               | ``TRACING``                                | ``0``           | Enable tracing. Can be ``0`` or ``1``.                                                                                                                                           |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--workers``               | ``GUNICORN_WORKERS``                       | ``1``           | Number of Gunicorn workers for handling requests.                                                                                                                                |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--threads``               | ``GUNICORN_THREADS``                       | ``1``           | Number of threads to run per Gunicorn worker.                                                                                                                                    |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--max-requests``          | ``GUNICORN_MAX_REQUESTS``                  | ``0``           | Maximum number of requests gunicorn worker will process before restarting.                                                                                                       |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--max-requests-jitter``   | ``GUNICORN_MAX_REQUESTS_JITTER``           | ``0``           | Maximum random jitter to add to max-requests.                                                                                                                                    |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--keepalive``             | ``GUNICORN_KEEPALIVE``                     | ``2``           | The number of seconds to wait for requests on a Keep-Alive connection.                                                                                                           |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--access-log``            | ``GUNICORN_ACCESS_LOG``                    | ``false``       | Enable gunicorn access log.                                                                                                                                                      |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--pidfile``               | N/A                                        | None            | A file path to use for the Gunicorn PID file.                                                                                                                                    |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| ``--single-threaded``       | ``FLASK_SINGLE_THREADED``                  | ``0``           | Force the Flask app to run single-threaded. Also applies to Gunicorn. Can be ``0`` or ``1``.                                                                                     |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| N/A                         | ``FILTER_METRICS_ACCESS_LOGS``             | ``not debug``   | Filter out logs related to Prometheus accessing the metrics port. By default enabled in production and disabled in debug mode.                                                   |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| N/A                         | ``PREDICTIVE_UNIT_METRICS_ENDPOINT``       | ``/metrics``    | Endpoint name for Prometheus metrics. In k8s deployment default is ``/prometheus``.                                                                                              |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| N/A                         | ``PAYLOAD_PASSTHROUGH``                    | ``false``       | Skip decoding of payloads.                                                                                                                                                       |
+-----------------------------+--------------------------------------------+-----------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+

Running Processes
-----------------

The total processes started for the python server will be as follows:

* REST server. A Gunicorn server which will have 1 master process and N worker processes where N is defined by the environment variable GUNICORN_WORKERS.
* gRPC server. A master process and N worker processes where N is defined by the environment variable GRPC_WORKERS.
* Metrics server. A Gunicorn server with 1 master and 1 worker.
* Metrics collector. A single process that collects the metrics across workers.

 
