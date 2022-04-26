# Servers

By default Seldon installs two server farms using MLServer and Triton. By default these will be 1 replica each. Models are scheduled onto servers based on the server's resources and whether the capabilities of the server matches the requirements specified in the Model request. For example:

```{literalinclude} ../../../../samples/models/sklearn-iris-gs.yaml
:language: yaml
```

This model specifies the requirement `sklearn`

There is a default capabilities for each server as follows:

* MLServer
  ```{literalinclude} ../../../../operator/config/serverconfigs/mlserver.yaml
  :language: yaml
  :start-after: SELDON_SERVER_CAPABILITIES
  :end-before: SELDON_OVERCOMMIT_PERCENTAGE
* Triton
  ```{literalinclude} ../../../../operator/config/serverconfigs/triton.yaml
  :language: yaml
  :start-after: SELDON_SERVER_CAPABILITIES
  :end-before: SELDON_OVERCOMMIT_PERCENTAGE
  ```

## Autoscaling of Servers

This is in the roadmap.


