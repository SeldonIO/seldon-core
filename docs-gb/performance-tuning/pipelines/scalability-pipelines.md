# Pipeline Scalability Guide

Core 2 supports full **horizontal scaling** for the **dataflow engine**, **model gateway**, and **pipeline gateway**. Each service automatically distributes pipelines or models across replicas using consistent hashing, so you don’t need to manually assign workloads.

This guide explains **how scaling works**, **what configuration controls it**, and **what happens when replicas or pipelines/models change**.

## 1. How scaling works (at a glance)

| Component            | What it scales with                                                       | Max replicas used                                                 |
|----------------------|---------------------------------------------------------------------------|-------------------------------------------------------------------|
| **Dataflow engine**  | `#pipelines × #Kafka partitions` (capped by replicas)                     | `min(replicas, pipelines × partitions)`                           |
| **Model gateway**    | `#models × #Kafka partitions` (capped by replicas and maxNumConsumers)    | `min(replicas, min(models, maxNumConsumers) × partitions)`        |
| **Pipeline gateway** | `#pipelines × #Kafka partitions` (capped by replicas and maxNumConsumers) | `min(replicas, min(pipelines, maxNumConsumers) × partitions)`     |


Each pipeline/model **is loaded only on a subset of replicas**, and **automatically rebalanced** when:
1. You **scale replicas up/down**
2. You **deploy or delete pipelines / models**

You **do not** need to manually assign work — it’s handled automatically.

## 2. Scaling the dataflow engine

**Dataflow engine** is responsible for **executing pipeline logic**. Core 2 now supports running **multiple pipelines in parallel across multiple dataflow engine replicas**.

### 2.1. What controls scaling?

You control scaling using:

| Config                             | Location        | Purpose                                                        |
|------------------------------------|-----------------|----------------------------------------------------------------|
| `spec.replicas`                    | `SeldonRuntime` | Maximum number of dataflow engine instances                    | 
| `kafkaConfig.topics.numPartitions` | `SeldonConfig`  | Determines max replication per pipeline (`#Kafka partitions`)  |

### 2.2. How many replicas will actually be used?

Dataflow engine replicas are dynamically adjusted based on **number of pipelines deployed** and **Kafka partitions**. The final number of dataflow engine replicas is given by:

$\text{FinalReplicaCount} = \min(\text{spec.replicas},\ \text{pipelines} \times \text{partitions})$

**Example**

| Pipelines deployed | Kafka partitions	 | spec.replicas	 | Final dataflow replicas used      |
|--------------------|-------------------|-------------------|-----------------------------------|
| 3	                 | 4                 | 9	             | `min(9, 3 x 4 = 12)` → 9 replicas |
| 2	                 | 4	             | 9         	     | `min(9, 2 x 4 = 8)` → 8 replicas  |
| 1                  | 4                 | 9	             | `min(9. 1 x 4 = 4)` → 4 replicas  |

**Note**: Unused replicas are automatically scaled down. As more pipelines are added, dataflow engine automatically scales up, capped by the maximum number of replicas.

### 2.3. How are pipeline assigned to replicas?

- Core 2 uses **consistent hashing** to distribute pipelines evenly across dataflow replicas. This ensures a **balanced workload**, but it does *not* guarantee a perfect one-to-one mapping.
    - Even if the number of replicas equals `pipelines × partitions`, some replicas may host **multiple pipelines** while others may **host none**, depending on how the hash falls. In practice, the distribution is **statistically uniform**, not strictly exact.
- Each pipeline is **replicated across multiple dataflow engines** (up to number of Kafka partitions).
- When instances are added or removed, **pipelines are automatically rebalanced**.

**Note:** This process is handled internally by Core 2, so no manual intervention is needed.

### 2.4. Loading/unloading of the pipelines from dataflow engine

- Loading/unloading of the pipeline from the dataflow engine is performed when the model CR is loaded/unloaded.
- The scheduler confirms whether the loading/unloading was performed successfully through the `Pipeline` status under the CR.

Rebalancing happens in the background — you don’t need to intervene.

**Note:** For pipelines, `Pipeline` ready status must be satisfied in order for the pipeline to be marked ready.

## 3. Scaling the model gateway

The **model gateway** is responsible for routing inference requests to models when used inside pipelines. Like the **dataflow engine**, it scales dynamically based on *how many models are deployed*.

### 3.1. What controls scaling?

| Config	                         | Location	                                                | Purpose                                                    |
|------------------------------------|----------------------------------------------------------|------------------------------------------------------------|
| `spec.replcias`	                 | `SeldonRuntime`	                                        | Maximum number of model gateway instances                  |
| `kafkaConfig.topics.numPartitions` | `SeldonConfig`	                                        | Determines max replication per model (`#Kafka partitions`) |
| `maxNumConsumers`	                 | `SeldonConfig` - model gateway enc var (default: 100)	| Caps how many distinct consumer groups can exist           |

### 3.2. How many replicas will actually be used?

Model gateway replicas are dynamically adjusted based on **number of models deployed**, **Kafka partitions**, and **maxNumConsumers**. The final number of model gateway replicas is given by: 

$\text{FinalReplicaCount} = \min(\text{spec.replicas},\ \min(\text{models}, \text{maxNumConsumers}) \times \text{partitions})$

**Example**

| Models Deployed |	Kafka Partitions | spec.replicas | maxNumConsumers | Final model gateway replicas                       |
|-----------------|------------------|---------------|-----------------|----------------------------------------------------|
| 5	              | 4	             | 20	         | 100	           | `min(20, min(5, 100) x 4 = 20) = 20` → 20 replicas |
| 1	              | 4	             | 20	         | 100	           | `min(20, min(1, 100) x 4 = 4) = 4` → 4 replicas    |

**Note:** If you remove models, the model gateway automatically scales down, and if we add model, the model gateway automatically scales up, capped by the maximum number of replicas.

### 3.3. How are models assigned to replicas?

Model gateway doesn’t load every model on every replica but only on a subset of replicas. The same principle as for dataflow engine applies for model gateway (sharding through consistent hashing).

### 3.4. Loading/unloading of the models from model gateway

- Loading/unloading of the model from the model gateway is performed when the model CR is loaded/unloaded.
- The scheduler confirms whether the loading/unloading was performed successfully through the `ModelGw` status under the CR.

Rebalancing happens in the background — you don’t need to intervene.

**Note:** `ModelGw` status does not represent a condition for the model to be available. If the loading was successful on the dedicated servers, the model itself is ready for inference. 

- The `ModelGw` status becomes relevant for pipeline and whether the end user wants to perform inference via the async path (i.e., writing the requests in the model input topic and reading the responses from the model output topic from Kafka).
- In the context of pipelines, the `ModelReady` status becomes a conjunction on whether the model is available on servers and if the model has been loaded successfully on the model gateway.

## 4. Scaling the pipeline gateway

The **pipeline gateway** is responsible for writing the requests in the input topic of the pipeline, and wait for the response on the output topic. Like dataflow engine and model gateway, pipeline gateway can scale horizontally.

### 4.1. What Controls Scaling?

| Config	                        | Location	                                               | Purpose                                                      |
|-----------------------------------|----------------------------------------------------------|--------------------------------------------------------------|
| `spec.replcias`	                | `SeldonRuntime`	                                       | Maximum number of pipeline gateway instances                 |
| `kafkaConfig.topics.numPartitions`| `SeldonConfig`	                                       | Determines max replication per pipeline (`# Kafka partitions`)  |
| `maxNumConsumers`	                | `SeldonConfig` - Pipeline gateway enc var (default: 100) | Caps how many distinct consumer groups can exist             |

### 4.2. How many replicas will actually be used?

Pipeline gateway replicas are dynamically adjusted based on **number of pipelines deployed**, **Kafka partitions**, and **maxNumConsumers**. The final number of pipeline gateway replicas is given by: 

$\text{FinalReplicaCount} = \min(\text{spec.replicas},\ \min(\text{pipelines}, \text{maxNumConsumers}) \times \text{partitions})$

**Example**

| Pipelines Deployed | Kafka Partitions	| spec.replicas	| maxNumConsumers	| Final pipeline gateway replicas                    |
|--------------------|------------------|---------------|-------------------|----------------------------------------------------|   
| 8	                 | 4	            | 10	        | 100	            | `min(10, min(8, 100) x 4 = 32) = 10` → 10 replicas |
| 2	                 | 4	            | 10	        | 100	            | `min(10, min(2, 100) x 4 = 8) = 8` → 8 replicas    | 
| 1	                 | 4	            | 10	        | 100	            | `min(10, min(1, 100) x 4 = 4) = 4`→ 4 replicas     |


**Note:** Similarly to dataflow engine, pipeline gateway scales up and down as pipeline are added and removed.

### 4.3. How are pipeline assigned to replicas?

Pipeline gateway doesn’t load every pipeline on every replica but only on a subset of replicas. The same principle as for dataflow engine and model gateway applies for pipeline gateway (sharding through consistent hashing).

### 4.4. Loading/unloading of the models from Pipeline Gateway

- Loading/unloading of the pipeline from the pipeline gateway is performed when the model CR is loaded/unloaded.
- The scheduler confirms whether the loading/unloading was performed successfully through the `PipelineGw` status under the CR.

Analogous with the previous services, rebalancing happens in the background — you don’t need to intervene.

**Note:** For pipelines, `PipelineGw` ready status must be satisfied in order for the pipeline to be marked ready.