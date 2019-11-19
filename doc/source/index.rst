===========
Seldon Core
===========

.. image:: seldon.png
   :alt: Seldon logo
   :align: center

Goals
-----

Seldon Core is an open source platform for deploying machine learning models on a Kubernetes cluster.

 * Deploy machine learning models in the cloud or on-premise.
 * Get metrics and ensure proper governance and compliance for your running machine learning models.
 * Create powerful inference graphs made up of multiple components.
 * Provide a consistent serving layer for models built using heterogeneous ML toolkits.

.. toctree::
   :maxdepth: 1
   :caption: Getting Started

   Simple Model Serving  <servers/overview.md>
   Advanced Custom Serving <workflow/README.md>

.. toctree::
   :maxdepth: 1
   :caption: Workflow
  
   Install  <workflow/install.md>
   Wrap your model <wrappers/README.md>   
   Create your inference graph <graph/inference-graph.md>
   Deploy your model  <workflow/deploying.md>
   Test your model <workflow/api-testing.md>
   Serve requests  <workflow/serving.md>
   Troubleshooting guide <workflow/troubleshooting.md>
   Usage reporting <workflow/usage-reporting.md>
   Upgrading <reference/upgrading.md>

.. toctree::
   :maxdepth: 1
   :caption: Servers
	     
   Inference Servers Overview <servers/overview.md>
   MLflow Server <servers/mlflow.md>
   SKLearn server <servers/sklearn.md>
   Tensorflow Serving <servers/tensorflow.md>
   XGBoost server <servers/xgboost.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Wrappers

   Python Models <python/index>   	     
   Java Models <java/README.md>
   R Models <R/README.md>
   NodeJS Models <nodejs/README.md>
   Custom Metrics <analytics/custom_metrics.md>
   Logging & Log Level <analytics/log_level.md>
	     
.. toctree::
   :maxdepth: 1
   :caption: Inference Graphs

   Distributed Tracing <graph/distributed-tracing.md>
   Annotation-based Configuration <graph/annotations.md>
   Private Docker Registry <graph/private_registries.md>
   Service Orchestrator <graph/svcorch.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Ingress

   Ambassador Ingress <ingress/ambassador.md>
   Istio Ingress <ingress/istio.md>
   Seldon OAuth Gateway <ingress/seldon.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Deployment Options

   Helm Charts <graph/helm_charts.md>	     
   Grafana Analytics <analytics/analytics.md>
   Elastic Stack Logging <analytics/logging.md>
   Autoscaling <graph/autoscaling.md>
	     
.. toctree::
   :maxdepth: 1
   :caption: ML Compliance and Governance

   Model Explanations <analytics/explainers.md>
   Outlier Detection <analytics/outlier_detection.md>
   Routers (incl. Multi Armed Bandits)  <analytics/routers.md>   
   
.. toctree::
   :maxdepth: 1
   :caption: Examples

   Notebooks <examples/notebooks>
   Integrations <examples/integrations>

.. toctree::
   :maxdepth: 1
   :caption: Tutorials

   Articles/Blogs <tutorials/blogs>
   Videos <tutorials/videos>

.. toctree::
   :maxdepth: 1
   :caption: Reference

   General Availability <reference/ga.md>
   Python API reference <python/api/modules>	     
   Seldon Microservice API <reference/apis/internal-api.md>
   Seldon Orchestrator <reference/engine>
   Benchmarking <reference/benchmarking.md>
   Seldon Deployment CRD <reference/seldon-deployment.md>
   Prediction APIs <reference/apis/index>
   Seldon Core Helm Chart <reference/helm.md>
   Release Highlights <reference/release-highlights>
   Images <reference/images.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Developer

   Overview <developer/readme.md>
   Roadmap <developer/roadmap.md>
   Build using private repo <developer/build-using-private-repo.md>

