# RedHat Operator Release Steps

## Summary

We presently still use the v1beta1 CRD. At some point we need to convert to the v1 CRD. However this CRD is too large for operator-registry (it converts the CRD to a configmap and it hits configmap limit it seems). We therefore might need to move forward with just v1 version for the v1 CRD and remove v1alpha2 and v1alpha3 versions of the SeldonDeployment CRD. See https://github.com/operator-framework/operator-registry/issues/385

There are also fixes in crd and crd_v1 configs for https://github.com/kubernetes/kubernetes/issues/91395 under a patch called protocol.yaml

We also remove the `owned` versions for v1alpha2 and v1alpha3 using `hack/csv_hack.py` to fix community test lint failures. This maybe actually be an issue in `operator-courier verify`.

## Prerequisites

Install [operator-sdk](https://sdk.operatorframework.io/).

Tested on

```
operator-sdk version
operator-sdk version: "v1.2.0", commit: "215fc50b2d4acc7d92b36828f42d7d1ae212015c", kubernetes version: "v1.18.8", go version: "go1.15.3", GOOS: "linux", GOARCH: "amd64"
```


## Version Update

Update Makefile and change PREV_VERSION to previous version

Login to quay.io as seldon. Password in 1password. 

```bash
make update_openshift
```

Updated image should be available in quay.io (https://quay.io/signin)

![quay-seldon](quay-seldon.png)

If quay does not contain the previous version specified in the CSV these steps will fail. You will need add and push to quay all previous bundles or craft the previous bundle by hand and remove its previous to stop the chain of references.

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

```
COMMUNITY_OPERATORS_FOLDER=~/work/seldon-core/redhat/community-operators

cp -r packagemanifests/1.3.0 ${COMMUNITY_OPERATORS_FOLDER}/community-operators/seldon-operator
cp packagemanifests/seldon-operator.package.yaml ${COMMUNITY_OPERATORS_FOLDER}/community-operators/seldon-operator
```

Run tests

```
cd ${COMMUNITY_OPERATORS_FOLDER}
make operator.test KUBE_VER=""  OP_PATH=community-operators/seldon-operator
```

Add new folder and changed package yaml to a PR. Ensure you sign the commit.

```
git commit -s -m "Update Seldon Community Operator to 1.2.2"
```

Push and create PR.

Do the same for the upstream community operators

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

Install [opm](https://docs.openshift.com/container-platform/4.6/cli_reference/opm-cli.html#opm-cli). I used docker instead of podman to install.

```
opm version
Version: version.Version{OpmVersion:"1.12.3", GitCommit:"", BuildDate:"2020-09-18T09:16:12Z", GoOs:"linux", GoArch:"amd64"}
```

Will need to be run in release branch.

Create new package

```
make create_certified_bundle
```

```
make build_certified_bundle
```


Push all images to redhat. requires download of passwords from 1password to `~/.config/seldon/seldon-core/redhat-image-passwords.sh`

```
cd {project_base_folder}/marketplaces/redhat
python scan-images.py
```

After these are finished (approx 1.5 hours) you will need to manually publish images on https://connect.redhat.com/project/5892531/images


Create a new catalog for certified on quay for testing.

```
update_openshift_certified
```

Test as above for openshift but using the new catalog source for certified. 


Push bundle image to scanning and tests. Also needs passwords.

```
make bundle_certified_push
```

TODO: seems to be differences in `replaces` in csv which says `seldon-operastor` but the package is called `seldon-operator-certified`

Publish image for final step to release new version of operator.

## Prepare for next release

Update base config version for next release

```
make update_config
```

