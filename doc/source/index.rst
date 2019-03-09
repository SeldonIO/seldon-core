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

   Overview <getting_started/index>

.. toctree::
   :maxdepth: 1
   :caption: Examples

   examples/iris

.. toctree::
   :maxdepth: 2
   :caption: Wrappers

   Python <python/index>


.. toctree::
   :maxdepth: 1
   :caption: Reference

   Python API reference <python/api/modules>   
   Seldon Operator <reference/cluster-manager>
   
Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`
