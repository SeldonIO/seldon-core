## GPU usage with Seldon

### 1. Enabling GPUs workload
#### Triton
Configure per-model GPU use in `model-settings`. TODO: add a GPU-focused example, because current examples don’t show this.

#### MLServer
For built-in runtimes, GPU selection is framework-specific (e.g., MLflow, vLLM params). For custom runtimes, we don’t have formal docs yet, so refer to this notebook: https://github.com/SeldonIO/customer-success/blob/master/catalog/standard-IQ-sessions/custom-usage/mlserver-gpu_v1.ipynb.

---

### 2. GPU partitioning & scaling patterns
Ways to share or expand GPU capacity across models/servers.

#### MIG (Multi-Instance GPU)
MIG splits one physical GPU into multiple isolated slices. Pods can request a MIG slice instead of a whole device, and this is best for hard isolation.

#### Multiple GPUs per server
Use multiple GPUs per server when a single model benefits from model-parallelism or high throughput. We still need to clarify/verify how GPUs are exposed to the runtime, the scheduling semantics, and whether per-model pinning is supported. Do we have it?

#### Time-slicing / GPU sharing
Time-slicing requires an infra prerequisite: it must be enabled at the cluster/driver level first. In MLServer, put models in the same inference pool and sharing happens within that pool. In Triton, use `instance_groups` to control concurrency and sharing. This is best for bursty workloads where strict isolation isn’t needed.

---

### 3. When to use what
Use MIG when you want isolation and consistent performance per model. Use multiple GPUs when one model needs more than one GPU. Use time-slicing when you have many light/bursty models and want to maximize utilization.