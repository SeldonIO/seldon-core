==================================
Seldon Deployment Reference Types
==================================

A SeldonDeployment is defined as a custom resource definition within Kubernetes.

If you want to learn about more practical examples of the use of the SeldonDeployment types you can check the `Inference Graph Section on the Worfklow Documentation Section <../graph/inference-graph.html>`_.

Below is the 2nd half of our `seldondeployment_types.go <https://github.com/SeldonIO/seldon-core/blob/master/operator/apis/machinelearning.seldon.io/v1/seldondeployment_types.go>`_ file which contains the types that are used in the SeldonDeployment YAML files.

.. literalinclude:: ../../../operator/apis/machinelearning.seldon.io/v1/seldondeployment_types.go
   :language: go
   :lines: 203-


