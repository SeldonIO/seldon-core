# Seldon Core Helm Charts

Helm charts are published to our official repo.

## Core Charts

The core charts for installing Seldon Core are shown below.

.. toctree::
   :maxdepth: 1

   seldon-core-operator <../charts/seldon-core-operator>
   seldon-core-analytics <../charts/seldon-core-analytics>

For further details see [here](../workflow/install.md).

## Inference Graph Templates

A set of charts to provide example templates for creating particular inference graphs using Seldon Core

.. toctree::
   :maxdepth: 1

   seldon-single-model <../charts/seldon-single-model>
   seldon-abtest <../charts/seldon-abtest>
   seldon-mab <../charts/seldon-mab>
   seldon-od-model <../charts/seldon-od-model>
   seldon-od-transformer <../charts/seldon-od-transformer>

[A notebook with examples of using the above charts](https://github.com/SeldonIO/seldon-core/tree/master/notebooks/helm_examples.ipynb) is provided.

## Misc

.. toctree::
   :maxdepth: 1

   seldon-core-loadtesting <../charts/seldon-core-loadtesting>
