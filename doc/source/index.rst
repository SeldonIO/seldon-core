===========
Seldon Core
===========

An open source platform to deploy your machine learning models on Kubernetes at massive scale.

Overview
-----
Seldon core converts your ML models (Tensorflow, Pytorch, H2o, etc.) or language wrappers (Python, Java, etc.) into production REST/GRPC microservices.

Seldon handles scaling to thousands of production machine learning models and provides advanced machine learning capabilities out of the box including Advanced Metrics, Request Logging, Explainers, Outlier Detectors, A/B Tests, Canaries and more.

Quick Links
-----

* Read the `Seldon Core Documentation <https://docs.seldon.io/projects/seldon-core/en/latest/>`_
* Join our `community Slack <https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLTJlNjQ1NTE5Y2MzMWIwMGUzYjNmZGFjZjUxODU5Y2EyMDY0M2U3ZmRiYTBkOTRjMzZhZjA4NjJkNDkxZTA2YmU>`_ to ask any questions
* Get started with `Seldon Core Notebook Examples <https://docs.seldon.io/projects/seldon-core/en/latest/examples/notebooks.html>`_
* Join our fortnightly `online community calls <https://docs.seldon.io/projects/seldon-core/en/latest/developer/community.html>`_
* Learn how you can `start contributing <https://docs.seldon.io/projects/seldon-core/en/latest/developer/contributing.html>`_
* Check out `Blogs <https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/blogs.html>`_ that dive into Seldon Core components
* Watch some of the `Videos and Talks <https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/videos.html>`_ using Seldon Core

.. image:: ./images/seldon-core-high-level.jpg
   :alt: Seldon logo
   :align: center

Documentation Index
----

.. toctree::
   :maxdepth: 1
   :caption: Getting Started

   Overview <workflow/github-readme.rst>
   Quickstart Guide <workflow/quickstart.md>
   Install Seldon Core on Kubernetes <workflow/install.md>
   Join the Community <developer/community.md>

.. toctree::
   :maxdepth: 1
   :caption: Seldon Core Deep Dive
  
   Detailed Installation Parameters <reference/helm.rst>
   Pre-packaged Inreference Servers <servers/overview.md>
   Language Wrappers for Custom Models <wrappers/language_wrappers.md>
   Create your Inference Graph <graph/inference-graph.md>
   Deploy your Model  <workflow/deploying.md>
   Testing your Model Endpoints  <workflow/serving.md>
   Python Module and Client <python/index.rst>
   Troubleshooting guide <workflow/troubleshooting.md>
   Usage reporting <workflow/usage-reporting.md>
   Upgrading <reference/upgrading.md>
   Changelog <reference/changelog.rst>

.. toctree::
   :maxdepth: 1
   :caption: Pre-Packaged Inference Servers
	     
   MLflow Server <servers/mlflow.md>
   SKLearn server <servers/sklearn.md>
   Tensorflow Serving <servers/tensorflow.md>
   XGBoost server <servers/xgboost.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Language Wrappers (Production)

   Python Language Wrapper [Production] <python/index.rst>

.. toctree::
   :maxdepth: 1
   :caption: Incubating Projects

   Java Language Wrapper [Incubating] <java/README.md>
   R Language Wrapper [ALPHA] <R/README.md>
   NodeJS Language Wrapper [ALPHA] <nodejs/README.md>
   Go Language Wrapper [ALPHA] <go/go_wrapper_link.rst>
   Stream Processing with KNative <streaming/knative_eventing.md>

.. toctree::
   :maxdepth: 1
   :caption: Ingress

   Ambassador Ingress <ingress/ambassador.md>
   Istio Ingress <ingress/istio.md>

.. toctree::
   :maxdepth: 1
   :caption: Production

   Supported API Protocols <graph/protocols.md>
   CI/CD MLOps at Scale <analytics/cicd-mlops.md>
   Metrics with Prometheus <analytics/analytics.md>
   Payload Logging with ELK <analytics/logging.md>
   Distributed Tracing with Jaeger <graph/distributed-tracing.md>
   Replica Scaling  <graph/scaling.md>
   Custom Inference Servers <servers/custom.md>
      
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
   Articles/Blogs <tutorials/blogs>
   Videos <tutorials/videos>
   Podcasts <tutorials/podcasts>

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
   Seldon Deployment CRD <reference/seldon-deployment.md>   
   Service Orchestrator <graph/svcorch.md>
   Kubeflow <analytics/kubeflow.md>

.. toctree::
   :maxdepth: 1
   :caption: Developer

   Overview <developer/readme.md>
   Contributing to Seldon Core <developer/contributing.rst>
   End to End Tests <developer/e2e.rst>
   Roadmap <developer/roadmap.md>
   Build using private repo <developer/build-using-private-repo.md>

