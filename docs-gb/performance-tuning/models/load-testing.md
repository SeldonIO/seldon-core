# Models

## Load Testing

Before looking to make changes to improve latency or throughput of your models, it is important to undergo load testing to understand the existing performance characteristics of your model(s) when deployed onto the chosen inference server (MLServer or Triton). The goal of load testing should be to understand the performance behavior of deployed models in saturated (i.e at the maximum throughput a model replica can handle) and non-saturated regimes, then compare it with expected latency objectives.

The results here will also inform the setup of autoscaling parameters, with the target being to run each replica with some margin below the saturation throughput (say, by 10-20%) in order to ensure that latency does not degrade, and that there is sufficient capacity to absorb load for the time it takes for new inference server replicas (and model replicas) to become available.

<aside>
💡
When testing latency, it is recommended to track different *percentiles* of latency (e.g. p50, p90, p95, p99). Choose percentiles based on the needs of your application - higher percentiles will help understand the variability of performance across requests.

</aside>

### Determining the Load Saturation point

A first target of load testing should be determining the maximum throughput that one model replica is able *to sustain*. We’ll refer to this as the (single replica) **load saturation point**. Increasing the number of inference requests per second (RPS) beyond this point degrades latency due to queuing that occurs at the bottleneck (a model processing step, contention on a resource such as CPU, memory, network, etc).  

We recommend determining the load saturation point by running an open-loop load test that goes through a series of stages, each having a target RPS. A stage would first linearly ramp-up to its target RPS and then hold that RPS level constant for a set amount of time. The target RPS should monotonically increase between the stages.

<aside>
💡

In order to get reproducible load test results, it is recommended to run the inference servers corresponding to the models being load-tested on k8s nodes where other components are not concurrently using a large proportion of the shared resources (compute, memory, IO). 

</aside>

### Closed-loop vs. Open-loop mode

Typically, load testing tools generate load by creating a number of "virtual users" that send requests. Knowing the behaviour of those virtual users is critical when interpreting the results of the load test. Load tests can be set up in **closed-loop mode**, meaning that when each virtual user sends a request *and they wait* for a response before sending the next one. Alternatively, there is **open-loop mode,** where a variable number of users are instantiated in order to maintain a constant overall RPS. 

- When running in **closed-loop mode**, an undesired side-effect called coordinated omission appears: when the system gets overloaded and latency spikes up, the load test users effectively reduce the actual load on the system by sending requests less frequently
- When using **closed-loop mode** in load testing, be aware that reported latencies at a given throughput may be significantly smaller than what will be experienced in reality
- In contrast, an **open-loop load** tester would maintain a constant RPS, resulting in a more accurate representation of the latencies that will be experienced when running the model in production.
- You can refer to the documentation of your load testing tool (i.e [Locust](https://www.locust.cloud/blog/closed-vs-open-workload-models), [k6](https://grafana.com/docs/k6/latest/using-k6/scenarios/executors/)) for guidance on choosing the right workload model (open vs. closed-loop) based on your testing goals.

