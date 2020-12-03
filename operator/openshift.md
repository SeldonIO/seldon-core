# RedHat Operator Release Steps

## Summary

We presently still use the v1beta1 CRD. At some point we need to convert to the v1 CRD. However this CRD is too large for operator-registry (it converts the CRD to a configmap and it hits configmap limit it seems). We therefore might need to move forward with just v1 version for the v1 CRD and remove v1alpha2 and v1alpha3 versions of the SeldonDeployment CRD. See https://github.com/operator-framework/operator-registry/issues/385

There are also fixes in crd and crd_v1 configs for https://github.com/kubernetes/kubernetes/issues/91395 under a patch called protocol.yaml

We also remove the `owned` versions for v1alpha2 and v1alpha3 using `hack/csv_hack.py` to fix community test lint failures. This maybe actually be an issue in `operator-courier verify`.

## Prerequisites

### Operastor-SDK

Install [operator-sdk](https://sdk.operatorframework.io/).


```
operator-sdk version
operator-sdk version: "v1.2.0", commit: "215fc50b2d4acc7d92b36828f42d7d1ae212015c", kubernetes version: "v1.18.8", go version: "go1.15.3", GOOS: "linux", GOARCH: "amd64"
```


### OPM

Install [opm](https://docs.openshift.com/container-platform/4.6/cli_reference/opm-cli.html#opm-cli). I used docker instead of podman to install.

```
opm version
Version: version.Version{OpmVersion:"1.12.3", GitCommit:"", BuildDate:"2020-09-18T09:16:12Z", GoOs:"linux", GoArch:"amd64"}
```

### Operator-Courier

See [https://github.com/operator-framework/operator-courier](here).

```
operator-courier -v
2.1.10
```

## Version Update

Update Makefile and change PREV_VERSION.

Login to quay.io as seldon. Password in 1password. 

```bash
make update_openshift
```

Updated image should be available in quay.io (https://quay.io/signin)

![quay-seldon](quay-seldon.png)


## Scorecard

```bash
kind create cluster
```

Run scorecard

```bash
make scorecard
```

## Tests

Run [kind cluster tests](./openshift/tests/README.md). k8s >= 1.16.

Run on an openshift cluster. Openshift >= 4.3.

## Community and Upstream Operators

Will need to be run in release branch

Create a fork of https://github.com/operator-framework/community-operators

Create a PR for community operator

Update the Makefile locally for 

```
COMMUNITY_OPERATORS_FOLDER=~/work/seldon-core/redhat/community-operators
```

### Community Operator

```
make update_community
```

Verify in community operator updated folder:

```
make operator.verify OP_PATH=community-operators/seldon-operator
```

Add new folder and changed package yaml to a PR. Ensure you sign the commit.

```
git commit -s -m "Update Seldon Community Operator to 1.2.2"
```

Push and create PR.

### Upstream Operator

```
COMMUNITY_OPERATORS_FOLDER=~/work/seldon-core/redhat/community-operators

cp -r packagemanifests/1.3.0 ${COMMUNITY_OPERATORS_FOLDER}/community-operators/seldon-operator
cp packagemanifests/seldon-operator.package.yaml ${COMMUNITY_OPERATORS_FOLDER}/community-operators/seldon-operator
```

Run tests

```
cd ${COMMUNITY_OPERATORS_FOLDER}
make operator.test KUBE_VER=""  OP_PATH=upstream-community-operators/seldon-operator
```

## Certified Operators


Create new package and push to quay for testing

```
make update_openshift_certified
```

Push all images to redhat. requires download of passwords from 1password to `~/.config/seldon/seldon-core/redhat-image-passwords.sh`

```
cd {project_base_folder}/marketplaces/redhat
python scan-images.py
```

After these are finished (approx 1.5 hours) you will need to manually publish images on https://connect.redhat.com/project/5892531/images

publish

 * https://connect.redhat.com/project/5912261/view
 * https://connect.redhat.com/project/5912271/view
 * https://connect.redhat.com/project/5912311/view
 * https://connect.redhat.com/project/5912301/view
 * https://connect.redhat.com/project/1366481/view
 * https://connect.redhat.com/project/1366491/view
 * https://connect.redhat.com/project/3977851/view
 * https://connect.redhat.com/project/3986991/view
 * https://connect.redhat.com/project/3987291/view
 * https://connect.redhat.com/project/3993461/view
 * https://connect.redhat.com/project/4035711/view


Test as above for openshift but using the new catalog source for certified. 


Push bundle image to scanning and tests. Also needs passwords.

```
make bundle_certified_push
```

This will start a test of the package in RedHat. Log on to check its success. If it fails you will need to manually delete in UI and build, tag and push a new version manually as you can't delete existing images. So a new tag will be needed which can be anything, e.g. 1.5.0-2, 1.5.0-3 etc.

## Prepare for next release

Update base config version for next release

```
make update_config
```

