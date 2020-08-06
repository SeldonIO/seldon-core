# Test on Vanilla Kubernetes with OLM

Follow https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#testing-operator-deployment-on-kubernetes

Using resource in this folder.

We set `startingCSV` to allow us to test auto upgrades by OLM. It should cycle through all operators from 1.1.0 onwards.
