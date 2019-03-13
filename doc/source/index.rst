===========
Seldon Core
===========

.. image:: seldon.png
   :alt: Seldon logo
   :align: center

Goals
-----

Seldon Core is a Kubernetes machine learning deployment platform. Its goals are:

 - Allow organisations to run and manage machine learning models built using any machine learning toolkit. Any model that can be run inside a Docker container can run in Seldon Core.
 - Provide a production ready machine learning deployment system on top of Kubernetes and integrating well with other Cloud Native tools.
 - Provide the tools to allow complex metrics, optimization and proper compliance of machine learning models in production.
   
   - Optimize your models using multi-armed bandit solvers
   - Run Outlier Detection models
   - Get alerts on Concept Drift
   - Provide black-box model explanations of running models
     
 - Provide APIs to allow business application to easily call your machine learning models to get predictions.
 - Handle full lifecycle management of the deployed model
   
   - Updating the runtime graph with no downtime
   - Scaling
   - Monitoring
   - Security


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
   usage reporting <workflow/usage-reporting.md>

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

   Helm Charts <graph/helm_charts>	     
   Ambassador Deployment <graph/ambassador.md>	     
   Example Grafana Analytics <analytics/analytics.md>
	     
.. toctree::
   :maxdepth: 1
   :caption: Metrics, Compliance, Governance

   Outlier Detection <analytics/outlier_detection>


   
.. toctree::
   :maxdepth: 1
   :caption: Examples

   Notebooks <examples/notebooks>
   Integrations <examples/integrations>

.. toctree::
   :maxdepth: 1
   :caption: Tutorials

   Articles/Blogs <tutorials/blogs.md>
   Videos <tutorials/videos.md>

.. toctree::
   :maxdepth: 1
   :caption: Reference

   Seldon Microservice API <reference/internal-api.md>
   Seldon Operator <reference/cluster-manager>
   Benchmarking <reference/benchmarking.md>
   Seldon Deployment CRD <reference/seldon-deployment.md>
   APIs <reference/apis/index>
   Seldon Core Helm Chart <reference/helm.md>
   Release Highlights <reference/release-highlights>

   
.. toctree::
   :maxdepth: 1
   :caption: Developer

   Roadmap <developer/roadmap.md>
   Overview <developer/readme.md>
   Build using private repo <developer/build-using-private-repo.md>

   
Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`
