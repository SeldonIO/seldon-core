---
description: Learn about the Seldon Agent API that enables communication between the Seldon Scheduler and Agent components for model management and request handling in Core 2.
---

# Agent API

This API is for communication between the Seldon Scheduler and the Seldon Agent which runs next to each inference server and manages the loading and unloading of models onto the server as well as acting as a reverse proxy in the data plane for handling requests to the inference server.

## Proto Definition

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/apis/mlops/agent/agent.proto" %}
