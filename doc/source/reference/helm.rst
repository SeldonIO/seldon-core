============
Advanced Helm Chart Configuration
============

Seldon Core Operator Chart Configuration
-----

The main file to install Seldon Core is the [seldon-core-operator Helm chart]().

Below you can find the values.yaml file of the helm chart, which are basically all the values that you can configure in your installation by using the `set` flag in the format `--set value.path=YOUR_VALUE`.

.. literalinclude:: ../../../helm-charts/seldon-core-operator/values.yaml
   :language: yaml
   :emphasize-lines: 12,15-18
   :linenos:

