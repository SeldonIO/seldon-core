---
name: Bug report
about: Create a bug report to help us improve
title: ''
labels: bug
assignees: ''

---

<!-- Welcome and thank you for helping us make Seldon Core better!

To help us address your issue, please provide us as much of the information requested below as possble. Thanks! -->


## Describe the bug
<!-- A clear and concise description of what the bug is. -->


## To reproduce
<!-- Steps required to reproduce the issue. For example:
1. define model ...
2. build image ...  (especially what wrapper version is used)
3. deploy ...
-->

## Expected behaviour
<!-- A clear and concise description of what you expected to happen. -->


## Environment
<!-- Description of environment -->

<!-- You Can fill it manually or paste the output of the command below:

* Cloud Provider: [e.g. GKE, AWS, Bare Metal, Kind, Minikube]
* Kubernetes Cluster Version [Output of `kubectl version`] 
* Deployed Seldon System Images: [Output of `kubectl get --namespace seldon-system deploy seldon-controller-manager -o yaml  | grep seldonio`]

Alternatively run `echo "#### Kubernetes version:\n $(kubectl version) \n\n#### Seldon Images:\n$(kubectl get --namespace seldon-system deploy seldon-controller-manager -o yaml  | grep seldonio)"`
-->

## Model Details <!-- If the issue is with your deployed model you can also provide the following for fulll insights -->
* Images of your model: [Output of: `kubectl get seldondeployment -n <yourmodelnamespace> <seldondepname> -o yaml | grep image:` where `<yourmodelnamespace>`]
* Logs of your model: [You can get the logs of your model by running `kubectl logs -n <yourmodelnamespace> <seldonpodname> <container>`]
