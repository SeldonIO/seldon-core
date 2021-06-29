# Overview

The setup of the nodes include the following

# General Node Pool

This is a node pool that is used for general processing, including the release build, the integration tests, etc.

There is a benchmarking node pool with the following requirements:
* taints: job-type=benchmark:NoSchedule

The command used to create it was the following:

```
gcloud container node-pools create general-pipelines-pool --zone=us-central1-a --cluster=tf-jx-working-weevil --node-taints=job-type=general:NoSchedule --enable-autoscaling --max-nodes=3 --min-nodes=0 --num-nodes=0 --machine-type=e2-standard-8  --disk-size=1000GB
```

It is possible to create pipelines that reference this job by using:

```
    nodeSelector:
      cloud.google.com/gke-nodepool: general-pipelines-pool
    tolerations:
    - key: job-type
      operator: Equals
      value: general
      effect: NoSchedule
```



# Benchmark Node Pool

This is the node pool that is used specifically for benchmarking tasks, where only 1 benchmark task would fit a single node.

There is a benchmarking node pool with the following requirements:
* taints: job-type=benchmark:NoSchedule

The command used to create it was the following:

```
gcloud container node-pools create benchmark-pipelines-pool --zone=us-central1-a --cluster=tf-jx-working-weevil --node-taints=job-type=benchmark:NoSchedule --enable-autoscaling --max-nodes=1 --min-nodes=0 --num-nodes=0 --machine-type=e2-standard-8  --disk-size=1000GB
```

It is possible to create pipelines that reference this job by using:

```
    nodeSelector:
      cloud.google.com/gke-nodepool: benchmark-pipelines-pool
    tolerations:
    - key: job-type
      operator: Equals
      value: benchmark
      effect: NoSchedule
```

