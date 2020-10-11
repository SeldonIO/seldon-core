# RedHat Operator Release Steps

Login to quay.io/seldon/seldon-operator

```bash
make recreate_bundle
make update_packagemanifests
make create_bundle_image
make push_bundle_image
make validate_bundle_image
make opm_index
make opm_push
```

## Scorecard

```bash
kind create cluster
```

Run scorecard

```bash
make scorecard
```

## Tests

Run on a kind cluster tests [./openshift/tests/README]. k8s >= 1.16.

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

Do the same for the upstream communit operators

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


Test as above for openshift but using the new catalog source. TODO: Needs creation of this catalogsource file in tests folder.


Push bundle image to scanning and tests. Also needs passwords.

```
make bundle_certified_push:
```


Publish image for final step to release new version of operator.
