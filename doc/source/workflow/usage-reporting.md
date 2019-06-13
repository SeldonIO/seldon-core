## Usage Reporting with Spartakus

An important part of the development process is to better understand the real user environment that the application will run in.

We provide an option to use an anonymous metrics collection tool provided by the Kubernetes project called [Spartakus](https://github.com/kubernetes-incubator/spartakus).

### Enable Usage Reporting

To help support the development of seldon-core, the voluntary reporting of usage data can be enabled whenever the "seldon-core-operator" helm chart is used  by setting the "--set usageMetrics.enabled=true" option.

```bash
helm install seldon-core-operator --name seldon-core \
        --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

The information that is reported is anonymous and only contains some information about each node in the cluster, including OS version, kubelet version, docker version, and CPU and memory capacity.

An example of what's reported:
```json
{
    "clusterID": "846db7e9-c861-43d7-8d08-31578af59878",
    "extensions": [
        {
            "name": "seldon-core-version",
            "value": "0.1.5"
        }
    ],
    "masterVersion": "v1.9.3-gke.0",
    "nodes": [
        {
            "architecture": "amd64",
            "capacity": [
                {
                    "resource": "cpu",
                    "value": "4"
                },
                {
                    "resource": "memory",
                    "value": "15405960Ki"
                },
                {
                    "resource": "pods",
                    "value": "110"
                }
            ],
            "cloudProvider": "gce",
            "containerRuntimeVersion": "docker://17.3.2",
            "id": "33082e677f61a199c195553e52bbd65a",
            "kernelVersion": "4.4.111+",
            "kubeletVersion": "v1.9.3-gke.0",
            "operatingSystem": "linux",
            "osImage": "Container-Optimized OS from Google"
        }
    ],
    "timestamp": "1522059083",
    "version": "v1.0.0-5d3377f6946c3ce9159cc9e7589cfbf1de26e0df"
}
```

### Disable Usage Reporting

Reporting of usage data is disabled by default, just use "seldon-core-operator" as normal.

```bash
helm install seldon-core-operator --name seldon-core \
        --repo https://storage.googleapis.com/seldon-charts
```

