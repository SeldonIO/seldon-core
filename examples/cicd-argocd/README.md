## Introduction

This example uses the following components to setup a demo "gitops" pipleline thats deploys a dummy ML model.

* Jenkins
* Argo
* Argocd
* Github (remote repos)
* Git (local repos)
* Local Docker Registry
* Seldon Core

![alt text](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/examples/cicd-argocd/cicd-demo.png "Seldon Code CICD demo")

## Setup

* directory for the scripts
```
seldon-core/examples/cicd-argocd
```
* Prepare the source and manifest Git repos
```
# Fork the following git repos to your Github account.
https://github.com/seldonio/cicd-demo-model-source-files
https://github.com/seldonio/cicd-demo-k8s-manifest-files

# Then clone the forks locally from your Github account.
cd /some/path/
git clone https://github.com/<your github id>/cicd-demo-model-source-files
git clone https://github.com/<your github id>/cicd-demo-k8s-manifest-files
```
* Create cluster, at least Kubernetes 10

* Have "helm", "argo", "argocd" installed

* (if gcp) create-user-cluster-admin-binding
```
kubectl create clusterrolebinding user-cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
```
* Create "settings.sh" using "settings.sh.example"

* Install helm
```
./seldon-core/install-helm
```
* Start all
```
./start-all
```
* Create tmux windows (Not from inside another tmux session)
```
./create-demo-tmux-session
```
* Manually test argo jobs
```
./argo/wf-run-build-image-and-push M01
./argo/wf-run-build-image-and-push M02
```
* Check images in the private registry
```
./k8s-local-docker-registry/registry-images-list
./k8s-local-docker-registry/registry-tags-list "gsunner/simple-model"
```
* Setup Jenkins
```
# get initial browser login details, and use top login
./jenkins/get-jenkins-browser-login

IMPORTANT: fix any plugin issues, eg. update pipeline-job plugin if necessary and Reboot

Install "Github" jenkins plugin
Manage jenkins->Manage Plugins->Available
    Find the "Github" plugin
    Install without restart

# setup security to use "Jenkins’ own user database"
Manage Jenkins->Configure Global Security
        - select "Jenkins’ own user database"
        - Make sure "Allow users to sign up" is unchecked
        - save

Jenkins will ask to "Create First Admin User"
            - use the JENKINS_USER_NAME and JENKINS_USER_PASSWORD in the "settings.sh" file
```
* Import Jenkins jobs
```
 ./jenkins/import-jobs
```
* Run the jenkins jobs to test

* Setup argocd
```
# get cmd-line login details, and use to login
./argocd/get-argocd-cmd-line-login

# add current cluster to list
./argocd/argocd-cluster-add
./argocd/argocd-cluster-list

# create app
./argocd/argocd-app-create

# get browser login details, and use to login
./argocd/get-argocd-browser-login
```
* Create Github Webhooks
```
# For CI, add webhook to "cicd-demo-model-source-files" repo
# get the webhook details to use
./jenkins/get-jenkins-github-webhook-details

# For CD, add webhook to "cicd-demo-k8s-manifest-files" repo
# get the webhook details to use
./argocd/get-argocd-github-webhook-details
```


## Clean Up

* Remove webhooks
```
# For CI, remove webhook from "cicd-demo-model-source-files" repo
# For CD, remove webhook from "cicd-demo-k8s-manifest-files" repo
```
* Stop all
```
./stop-all
```
* Remove helm
```
./seldon-core/remove-helm
```

