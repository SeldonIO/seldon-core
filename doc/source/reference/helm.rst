
=================================
Advanced Helm Chart Configuration
=================================

Seldon Core Operator Chart Configuration
----------------------------------------

This page provides a detailed overview of the installation parameters available for the Seldon Core installation when using Helm 3.x. The high level workflows to install Seldon Core can be found in the `Installation Page <../workflow/install.md>`_.

Below you can find the `values.yaml` file of the `seldon-core-operator Helm chart <https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-operator>`_, which contains basically all the values that you can configure in your installation by using the `set` flag in the format `--set value.path=YOUR_VALUE`.

The file has been written to be self documented, and has information on all the core parameters. Further information is referenced in the file to specific documentation pages.

.. literalinclude:: ../../../helm-charts/seldon-core-operator/values.yaml
   :language: yaml
   :linenos:

