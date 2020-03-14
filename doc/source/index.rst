===========
Seldon Core
===========

An open source platform to deploy your machine learning models on Kubernetes at massive scale.

.. image:: seldon-logo-small.png
   :alt: Seldon logo
   :align: center

Overview
-----
Seldon core converts your ML models (Tensorflow, Pytorch, H2o, etc.) or language wrappers (Python, Java, etc.) into production REST/GRPC microservices.

Seldon handles scaling to thousands of production machine learning models and provides advanced machine learning capabilities out of the box including Advanced Metrics, Request Logging, Explainers, Outlier Detectors, A/B Tests, Canaries and more.

Quick Links
-----

* Read the `Seldon Core Documentation <https://docs.seldon.io/projects/seldon-core/en/latest/>`_
* Join our `community Slack <https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLTJlNjQ1NTE5Y2MzMWIwMGUzYjNmZGFjZjUxODU5Y2EyMDY0M2U3ZmRiYTBkOTRjMzZhZjA4NjJkNDkxZTA2YmU>`_ to ask any questions
* Get started with `Seldon Core Notebook Examples <https://docs.seldon.io/projects/seldon-core/en/latest/examples/notebooks.html>`_
* Join our fortnightly `online community calls <>`_
* Learn how you can `start contributing <https://docs.seldon.io/projects/seldon-core/en/latest/developer/contributing.html>`_
* Check out `Blogs <https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/blogs.html>`_ that dive into Seldon Core components
* Watch some of the `Videos and Talks <https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/videos.html>`_ using Seldon Core

.. image:: ./images/seldon-core-high-level.jpg
   :alt: Seldon logo
   :align: center

.. toctree::
   :maxdepth: 1
   :caption: Getting Started

   Overview <workflows/overview.md>
   Quickstart Guide <workflows/quickstart.md>

.. toctree::
   :maxdepth: 1
   :caption: Workflow Deep Dive
  
   Install on Kubernetes <workflow/install.md>
   Wrap your model <wrappers/README.md>   
   Wrap your model with Pre-packaged Inreference Servers <servers/overview.md>
   Wrap your model with Custom Models with Language Wrappers <workflow/README.md>
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
   :caption: Language Wrappers

   Python Models <python/index>   	     
   Java Models <java/README.md>
   R Models <R/README.md>
   NodeJS Models <nodejs/README.md>
   Custom Metrics <analytics/custom_metrics.md>

.. toctree::
   :maxdepth: 1
   :caption: Ingress

   Ambassador Ingress <ingress/ambassador.md>
   Istio Ingress <ingress/istio.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Production

   Protocols <graph/protocols.md>	     
   Tracing <graph/distributed-tracing.md>
   Metrics <analytics/analytics.md>
   Payload Logging <analytics/logging.md>
   Autoscaling <graph/autoscaling.md>
      
.. toctree::
   :maxdepth: 1
   :caption: Advanced Inference

   Model Explanations <analytics/explainers.md>
   Outlier Detection <analytics/outlier_detection.md>
   Routers (incl. Multi Armed Bandits)  <analytics/routers.md>   
   
.. toctree::
   :maxdepth: 1
   :caption: Examples

   Notebooks <examples/notebooks>
   Integrations <examples/integrations>
   Articles/Blogs <tutorials/blogs>
   Videos <tutorials/videos>

.. toctree::
   :maxdepth: 1
   :caption: Reference

   Annotation-based Configuration <graph/annotations.md>   	     
   AWS Marketplace Install <reference/aws-mp-install.md>
   Benchmarking <reference/benchmarking.md>   
   General Availability <reference/ga.md>
   Helm Charts <graph/helm_charts.md>
   Images <reference/images.md>
   Logging & Log Level <analytics/log_level.md>   
   Private Docker Registry <graph/private_registries.md>   
   Prediction APIs <reference/apis/index>   
   Python API reference <python/api/modules>
   Release Highlights <reference/release-highlights>   
   Seldon Core Helm Chart <reference/helm.md>   
   Seldon Deployment CRD <reference/seldon-deployment.md>   
   Service Orchestrator <graph/svcorch.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Developer

   Overview <developer/readme.md>
   Contributing to Seldon Core <developer/contributing.rst>
   End to End Tests <developer/e2e.rst>
   Roadmap <developer/roadmap.md>
   Build using private repo <developer/build-using-private-repo.md>

