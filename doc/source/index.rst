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

   Overview <workflow/README.md>

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

.. toctree::
   :maxdepth: 1
   :caption: Wrappers

   Python Models <python/index>   	     
   Java Models <java/README.md>
   R Models <R/README.md>
   NodeJS Models <nodejs/README.md>
   Custom Metrics <analytics/custom_metrics.md>
	     
.. toctree::
   :maxdepth: 1
   :caption: Inference Graphs

   Distributed Tracing <graph/distributed-tracing.md>
   Annotation-based Configuration <graph/annotations.md>
   Private Docker Registry <graph/private_registries.md>


.. toctree::
   :maxdepth: 1
   :caption: Deployment Options

   Helm Charts <graph/helm_charts.md>	     
   Ambassador Deployment <graph/ambassador.md>	     
   Grafana Analytics <analytics/analytics.md>
	     
.. toctree::
   :maxdepth: 1
   :caption: ML Compliance and Governance

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

   Seldon Microservice API <reference/apis/internal-api.md>
   Seldon Operator <reference/cluster-manager>
   Benchmarking <reference/benchmarking.md>
   Seldon Deployment CRD <reference/seldon-deployment.md>
   Prediction APIs <reference/apis/index>
   Seldon Core Helm Chart <reference/helm.md>
   Release Highlights <reference/release-highlights>
   Images <reference/images.md>
   
.. toctree::
   :maxdepth: 1
   :caption: Developer

   Roadmap <developer/roadmap.md>
   Overview <developer/readme.md>
   Build using private repo <developer/build-using-private-repo.md>

