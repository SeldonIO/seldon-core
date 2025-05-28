# Introduction

In MLOps, system performance can be defined as how efficiently and effectively an ML model or application operates in a production environment. It is typically measured across several key dimensions, such as latency, throughput, scalability, and resource-efficiency. These factors are deeply connected: changes in configuration often result in tradeoffs, and should be carefully considered. Specifically, latency, throughput and resource usage can all impact each other, and the approach to optimizing system performance depends on the desired balance of these outcomes in order to ensure a positive end-user experience, while also minimising infrastructure costs.  

## High Level Approach

There are many different levers that can be considered to tune performance for ML systems deployed with Core 2, across infrastructure, models, inference execution, and the related configurations exposed by Core 2. When reasoning about the performance of an ML-based system deployed using Seldon, we recommend breaking down the problem by first understanding and tuning the performance of deployed **Models**, and then subsequently considering more complex **Pipelines** composed of those models (if applicable). For both models and pipelines, it is important to run tests to understand baseline performance characteristics, before making efforts to tune these through changes to models, infrastructure, or inference configurations. The recommended approach can be broken down as follows:

### Models

1. [**Load testing**](models/load-testing.md) to understand latency and throughput behaviour for one model replica
2. Tuning performance for models. This can be done via changes to:
    1. [**Infrastructure**](models/infrastructure-setup.md) - choosing the right hardware, and configurations in Core related to CPUs, GPUs and memory.
    2. [**Models**](models/inference.md#optimizing-the-model-artefact) - optimizing model artefacts in how they are structured, configured, stored. This can include model pruning, quantization, consideration of different model frameworks, and making sure that the model can achieve a high utilisation of the allocated resources.
    3. [**Inference**](models/inference.md#inference) - the way in which inference is executed. This can include the choice of communication protocols (REST, gRPC),  payload configuration, batching, and efficient  execution of concurrent requests.

### **Pipelines**

1. [**Testing Pipelines**](pipelines/testing-pipelines.md) to identify the critical path based on performance of underlying models
2. [**Core 2 Configuration**](pipelines/core-2-configuration.md) to optimize data-processing through pipelines  
