# Upgrading

## Upgrading from 2.5 - 2.6

Release 2.6 brings with it new custom resources `SeldonConfig` and `SeldonRuntime`, which provide a new way to install Seldon Core V2 in Kubernetes. Upgrading in the same namespace will cause downtime while the pods are being recreated. Alternatively  users can have an external service mesh or other means to be used over multiple namespaces to bring up the system in a new namespace and redeploy models before switch traffic between them.

If the new 2.6 charts are used to upgrade in an existing namespace models will eventually be redeloyed but there will be service downtime as the core components are redeployed.

