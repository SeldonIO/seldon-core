# Openshift Tests

## Local Kind Cluster

Based on [community operators docs](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#testing-operator-deployment-on-kubernetes). However, these docs are out of date and do not work with latest installs.

```
kind create cluster
```

Install latest OLM release.

```bash
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.16.1/crds.yaml
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.16.1/olm.yaml
```

For the marketplace you need a patched installed [pending in this PR](https://github.com/operator-framework/operator-marketplace/pull/342).

Temporary fix:

```bash
git clone git@github.com:cliveseldon/operator-marketplace.git fork-operator-marketplace
kubectl create -f fork-operator-marketplace/deploy/upstream/
```

Create the bundle image, check and push. Create opm_index, and push if not done so already.

```bash
cd ../..
make update_openshift
```

Create a catalog_source

```bash
kubectl create -f catalog-source.yaml
```

Check

```
kubectl get catalogsource seldon-core-catalog -n marketplace -o yaml
```

Create operator group

```bash
kubectl create -f operator-group.yaml
```

Create Subscription

Note a subscription that starts with csv <= 1.2.2 will fail on k8s>=1.18 as these earlier versions will not run on k8s 1.18.


```bash
kubectl create -f operator-subscription.yaml
```

This should create the seldon-controller manager. Once running you can test. It will be namespace only so will only manage sdeps in marketplace namespace.


## Openshift Cluster

Create catalog source

```bash
kubectl create -f catalog-source-openshift.yaml
```

Check

```
kubectl get catalogsource seldon-core-catalog -n openshift-marketplace -o yaml
```

At present need to create operator from Openshift UI.



The below does not seem to work.

Create operator group

```bash
kubectl create -f operator-group-openshift.yaml
```

Create Subscription

```bash
kubectl create -f operator-subscription-openshift.yaml
```

This should create the seldon-controller manager. Once running you can test. It will be namespace only so will only manage sdeps in marketplace namespace.




## Openshift Cluster Certified

Create catalog source

```bash
kubectl create -f catalog-source-openshift-certified.yaml
```

Check

```
kubectl get catalogsource seldon-core-catalog-certified -n openshift-marketplace -o yaml
```

At present need to create operator from Openshift UI.




