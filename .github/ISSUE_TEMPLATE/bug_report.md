---
name: Bug report
about: Create a bug report to help us improve
title: ''
labels: bug, triage
assignees: ''

---

<!-- Hi! Welcome to Seldon Core GitGub issue tracker!

To help us address your issue better, please, provide us as much of the information requested below as possible. Thanks! -->


## Describe the bug
<!-- A clear and concise description of what the bug is. -->


## To reproduce
<!-- Steps required to reproduce the issue.

Ideally, if we can reproduce the problem fully using a minimal model and `kind` k8s cluster
it will be easier for us to quickly and accurately  address the issue. -->

1. define model ...
2. build image ...  (especially what wrapper version is used)
3. deploy ...


## Expected behaviour
<!-- A clear and concise description of what you expected to happen. -->


## Environment
<!-- Description of environment -->

- Kubernetes Version: [e.g. 1.15]
- Cloud Provider: [e.g. GKE, AWS, Bare Metal, Kind, Minikube]
- Operating system and version: [eg. Mac OS 10.13.6, Ubuntu 18.04.3]
- Operator image: [e.g. seldon-core-operator:1.1.0]
- Orchestrator image:  [e.g. seldon-core-executor:1.1.0]
- Wrapper image: [e.g. seldon-core-s2i-python37:1.1.0]


<!-- Following commands can help you find required information
```
kubectl version
helm list -n seldon-system
kubectl get -n seldon-system deployments.apps seldon-controller-manager -o yaml  | grep seldonio
```
-->


## Additional context / Logs
<!-- Any additional information that you can provide us, especially logs of containers -->

<details><summary><code>My model logs:</code></summary><p>
<!-- Paste output of  'kubectl logs -n seldon ...'  in between the ticks below -->

```

```
</p></details>
