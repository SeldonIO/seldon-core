# Introduction

In MLOps, system performance can be defined as how efficiently and effectively an ML model or application operates in a production environment. It is typically measured across several key dimensions, such as latency, throughput, scalability, and resource-efficiency. These factors are deeply connected: changes in configuration often result in tradeoffs, and should be carefully considered. Specifically, latency, throughput and resource usage can all impact each other, and the approach to optimizing system performance depends on the desired balance of these outcomes in order to ensure a positive end-user experience, while also minimising infrastructure costs.  

## High Level Approach

There are many different levers that can be considered to tune performance for ML systems deployed with Core 2, across infrastructure, models, inference execution, and the related configurations exposed by Core 2. When reasoning about the performance of an ML-based system deployed using Seldon, we recommend breaking down the problem by first understanding and tuning the performance of deployed **Models**, and then subsequently considering more complex **Pipelines** composed of those models (if applicable). For both models and pipelines, it is important to run tests to understand baseline performance characteristics, before making efforts to tune these through changes to models, infrastructure, or inference configurations. The recommended approach can be broken down as follows:

### Models

1. [**Load testing**](https://www.notion.so/Performance-Tuning-Docs-1bc6a4c8852080ce94c4eb1dcd725b9f?pvs=21) to understand latency and throughput behaviour for one model replica
2. Tuning performance for models. This can be done via changes to:
    1. [**Infrastructure](https://www.notion.so/Performance-Tuning-Docs-1bc6a4c8852080ce94c4eb1dcd725b9f?pvs=21)** - choosing the right hardware, and configurations in Core related to CPUs, GPUs and memory.
    2. [**Models](https://www.notion.so/Performance-Tuning-Docs-1bc6a4c8852080ce94c4eb1dcd725b9f?pvs=21)** - optimizing model artefacts in how they are structured, configured, stored. This can include model pruning, quantization, consideration of different model frameworks, and making sure that the model can achieve a high utilisation of the allocated resources.
    3. [**Inference**](https://www.notion.so/Performance-Tuning-Docs-1bc6a4c8852080ce94c4eb1dcd725b9f?pvs=21) - the way in which inference is executed. This can include the choice of communication protocols (REST, gRPC),  payload configuration, batching, and efficient  execution of concurrent requests.

### **Pipelines**

1. [**Testing Pipelines**](https://www.notion.so/Performance-Tuning-Docs-1bc6a4c8852080ce94c4eb1dcd725b9f?pvs=21) to identify the critical path based on performance of underlying models
2. [**Core 2 Configuration**](https://www.notion.so/Performance-Tuning-Docs-1bc6a4c8852080ce94c4eb1dcd725b9f?pvs=21) to optimize data-processing through pipelines  
