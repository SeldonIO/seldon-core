# Seldon Operator for RedHat

## Resource Generation

For a new release:

Regenerate the core yaml from kustomize for the "lite" version of Core:

```
make generate-resources
```

Recreate the core yaml from these resources:

```
make deploy/operator.yaml
```

Create a new rule in the Makefile to generate the operator CSV from a previous version using the latest yaml. For 1.1.0 this is an initial rule based off a phony previous release 1.0.0.

```
make clean_1.1.0 deploy/olm-catalog/seldon-operator/1.1.0/seldon-operator.v1.1.0.clusterserviceversion.yaml
```

Check the OLM bundle. For this you need to have installed [operator-courier](https://github.com/operator-framework/operator-courier).

```
make operator-courier_ui_validate
```

Start a Kind cluster and test the Operator via the scorecard application. You need to have installed [operator-sdk](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md).

You will need a running cluster with kubectl authenticated and the latest images to test loaded onto the cluster.

```
make scorecard
```

The output should be simialr to:

```
INFO[0000] Using config file: /home/clive/work/seldon-core/fork-seldon-core/operator/seldon-operator/.osdk-scorecard.yaml 
basic:
	Spec Block Exists                   : pass
	Labels: 
		"necessity":"required"
		"suite":"basic"
		"test":"checkspectest"

	Status Block Exists                 : pass
	Labels: 
		"suite":"basic"
		"test":"checkstatustest"
		"necessity":"required"

	Writing into CRs has an effect      : pass
	Labels: 
		"necessity":"required"
		"suite":"basic"
		"test":"writingintocrshaseffecttest"

olm:
	Spec fields with descriptors        : pass
	Labels: 
		"test":"specdescriptorstest"
		"necessity":"required"
		"suite":"olm"

	Status fields with descriptors      : pass
	Labels: 
		"necessity":"required"
		"suite":"olm"
		"test":"statusdescriptorstest"

	Bundle Validation Test              : pass
	Labels: 
		"necessity":"required"
		"suite":"olm"
		"test":"bundlevalidationtest"
	Log:
		time="2020-03-18T15:39:42Z" level=info msg="loading Bundles" dir=deploy/olm-catalog/seldon-operator
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=seldon-operator load=bundles
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.2 load=bundles
time="2020-03-18T15:39:42Z" level=info msg="found csv, loading bundle" dir=deploy/olm-catalog/seldon-operator file=seldonoperator.v0.1.2.clusterserviceversion.yaml load=bundles
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.2 file=seldondeployments.machinelearning.seldon.io.crd.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.2 file=seldonoperator.v0.1.2.clusterserviceversion.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.3 load=bundles
time="2020-03-18T15:39:42Z" level=info msg="found csv, loading bundle" dir=deploy/olm-catalog/seldon-operator file=seldonoperator.v0.1.3.clusterserviceversion.yaml load=bundles
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.3 file=seldondeployments.machinelearning.seldon.io.crd.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.3 file=seldonoperator.v0.1.3.clusterserviceversion.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.4 load=bundles
time="2020-03-18T15:39:42Z" level=info msg="found csv, loading bundle" dir=deploy/olm-catalog/seldon-operator file=seldonoperator.v0.1.4.clusterserviceversion.yaml load=bundles
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.4 file=seldondeployments.machinelearning.seldon.io.crd.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.4 file=seldonoperator.v0.1.4.clusterserviceversion.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.5 load=bundles
time="2020-03-18T15:39:42Z" level=info msg="found csv, loading bundle" dir=deploy/olm-catalog/seldon-operator file=seldonoperator.v0.1.5.clusterserviceversion.yaml load=bundles
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.5 file=seldondeployments.machinelearning.seldon.io.crd.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/0.1.5 file=seldonoperator.v0.1.5.clusterserviceversion.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=1.0.0 load=bundles
time="2020-03-18T15:39:42Z" level=info msg="found csv, loading bundle" dir=deploy/olm-catalog/seldon-operator file=seldon-operator.v1.0.0.clusterserviceversion.yaml load=bundles
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/1.0.0 file=machinelearning.seldon.io_seldondeployment_crd.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/1.0.0 file=seldon-operator.v1.0.0.clusterserviceversion.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=1.1.0 load=bundles
time="2020-03-18T15:39:42Z" level=info msg="found csv, loading bundle" dir=deploy/olm-catalog/seldon-operator file=seldon-operator.v1.1.0.clusterserviceversion.yaml load=bundles
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/1.1.0 file=machinelearning.seldon.io_seldondeployment_crd.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg="loading bundle file" dir=deploy/olm-catalog/seldon-operator/1.1.0 file=seldon-operator.v1.1.0.clusterserviceversion.yaml load=bundle
time="2020-03-18T15:39:42Z" level=info msg="loading Packages and Entries" dir=deploy/olm-catalog/seldon-operator
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=seldon-operator load=package
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.2 load=package
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.3 load=package
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.4 load=package
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=0.1.5 load=package
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=1.0.0 load=package
time="2020-03-18T15:39:42Z" level=info msg=directory dir=deploy/olm-catalog/seldon-operator file=1.1.0 load=package


	Provided APIs have validation       : pass
	Labels: 
		"suite":"olm"
		"test":"crdshavevalidationtest"
		"necessity":"required"

	Owned CRDs have resources listed    : pass
	Labels: 
		"necessity":"required"
		"suite":"olm"
		"test":"crdshaveresourcestest"
```

Next step is to run manual tests on clusters.

## Standalone Resource Test

To run a manual test follow the instructions in [testing/standalone/README.md](./testing/standalone/README.md)

## Quay upload of Operator Bundle

Upload latest bundle to quay. You will need a quay token saved in QUAY_TOKEN environment variable. To get this you need to clone operator-courier and then use

```
./operator-courier/scripts/get-quay-token
```

For login details use 1password.

export token, e.g.

```
export QUAY_TOKEN="basic f2VsAG9uNmZydXsd12I2R0hPUCpwc3Vy"
```

Then run

```
make operator-courier_push 
```

You will need to delete any old application from quay before this will succeed.

Go to https://quay.io/application/seldon/seldon-operator?tab=settings and make it public.

## Vanilla K8S with OLM

Follow quay bundle upload above.

See [testing/k8s/README.md](testing/k8s/README.md).

## Openshift cluster

Follow quay bundle upload above.

See [testing/openshift/README.md](testing/openshift/README.md).

[Create an Openshift cluster on AWS](https://cloud.redhat.com/openshift/install/aws/installer-provisioned)

## Release Process onto RedHat channels

After a Seldon Release has taken place.

 * Update versions of images by running above updates and ensuring images referenced in the CSV are for the last release.
 * Run above tests to check all works ok.

For community operators:

 * Create a fork of https://github.com/operator-framework/community-operators and update Seldon and create a new PR

For certified operators:

 * Create UBI images for operator, executor and engine.
 * Create a new release on https://connect.redhat.com/content/seldon-core with
    * new ubi images
    * new bundle that refrences then. TODO: create bundle from community bundle.

For Marketplace:

 * Update Seldon on marketplace. TODO: add link.