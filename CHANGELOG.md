# Change Log

## [v0.1.6](https://github.com/SeldonIO/seldon-core/tree/v0.1.6) (2018-03-29)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.5...v0.1.6)

**Closed issues:**

- Support RBAC by default [\#126](https://github.com/SeldonIO/seldon-core/issues/126)
- Engine requires images to have versions [\#117](https://github.com/SeldonIO/seldon-core/issues/117)
- `hostPath` type volume gets mounted as `emptyDir` [\#116](https://github.com/SeldonIO/seldon-core/issues/116)
- Investigate OpenShift source-to-image for wrapping models [\#113](https://github.com/SeldonIO/seldon-core/issues/113)
- Add docs for analytics persistence [\#112](https://github.com/SeldonIO/seldon-core/issues/112)
- Issue in deployments of multiple models  [\#103](https://github.com/SeldonIO/seldon-core/issues/103)
- Missing dependencies in notebooks/kubectl\_demo\_minikube.ipynb [\#101](https://github.com/SeldonIO/seldon-core/issues/101)
- Add usage metrics collector [\#99](https://github.com/SeldonIO/seldon-core/issues/99)
- Running  test model on seldon core  [\#90](https://github.com/SeldonIO/seldon-core/issues/90)
- Deploying seldon models to multiple namespaces [\#89](https://github.com/SeldonIO/seldon-core/issues/89)
- Generate load tests analytics [\#58](https://github.com/SeldonIO/seldon-core/issues/58)
- Create wrapper for R models [\#23](https://github.com/SeldonIO/seldon-core/issues/23)

**Merged pull requests:**

- Rbac fixes [\#127](https://github.com/SeldonIO/seldon-core/pull/127) ([cliveseldon](https://github.com/cliveseldon))
- Anonymous usage metrics collection [\#125](https://github.com/SeldonIO/seldon-core/pull/125) ([gsunner](https://github.com/gsunner))
- R wrappers [\#124](https://github.com/SeldonIO/seldon-core/pull/124) ([cliveseldon](https://github.com/cliveseldon))
- Fix parsing of image version in engine [\#119](https://github.com/SeldonIO/seldon-core/pull/119) ([cliveseldon](https://github.com/cliveseldon))
- S2i examples [\#118](https://github.com/SeldonIO/seldon-core/pull/118) ([cliveseldon](https://github.com/cliveseldon))
- S2i integration [\#115](https://github.com/SeldonIO/seldon-core/pull/115) ([cliveseldon](https://github.com/cliveseldon))
- change benchmark notebook name [\#111](https://github.com/SeldonIO/seldon-core/pull/111) ([cliveseldon](https://github.com/cliveseldon))
- Benchmarking seldon-core [\#110](https://github.com/SeldonIO/seldon-core/pull/110) ([cliveseldon](https://github.com/cliveseldon))
- 55 outlier detection [\#105](https://github.com/SeldonIO/seldon-core/pull/105) ([Maximophone](https://github.com/Maximophone))
- Made notebooks compatible with python 3 [\#102](https://github.com/SeldonIO/seldon-core/pull/102) ([Maximophone](https://github.com/Maximophone))

## [v0.1.5](https://github.com/SeldonIO/seldon-core/tree/v0.1.5) (2018-02-19)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.4...v0.1.5)

**Closed issues:**

- Make CRD namespaced [\#95](https://github.com/SeldonIO/seldon-core/issues/95)
- Allow Helm deployment without API Front end [\#92](https://github.com/SeldonIO/seldon-core/issues/92)
- Support deployment of a Python 3 model [\#88](https://github.com/SeldonIO/seldon-core/issues/88)
- Create a Slack channel for project [\#43](https://github.com/SeldonIO/seldon-core/issues/43)

**Merged pull requests:**

- ksonnet updates for namespaces and RBAC [\#98](https://github.com/SeldonIO/seldon-core/pull/98) ([cliveseldon](https://github.com/cliveseldon))
- Handle namespaced deployments [\#97](https://github.com/SeldonIO/seldon-core/pull/97) ([cliveseldon](https://github.com/cliveseldon))
- 88 python 3 compatibility [\#94](https://github.com/SeldonIO/seldon-core/pull/94) ([Maximophone](https://github.com/Maximophone))
- allow apife to be optional in helm install [\#93](https://github.com/SeldonIO/seldon-core/pull/93) ([cliveseldon](https://github.com/cliveseldon))
- remove redundant assignment [\#87](https://github.com/SeldonIO/seldon-core/pull/87) ([mjlodge](https://github.com/mjlodge))

## [v0.1.4](https://github.com/SeldonIO/seldon-core/tree/v0.1.4) (2018-02-05)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.3...v0.1.4)

**Closed issues:**

- Create KSonnet prototypes for core [\#76](https://github.com/SeldonIO/seldon-core/issues/76)
- Add automatically generated README to wrapped models [\#57](https://github.com/SeldonIO/seldon-core/issues/57)

**Merged pull requests:**

- ksonnet notebook with Ambassador  [\#81](https://github.com/SeldonIO/seldon-core/pull/81) ([cliveseldon](https://github.com/cliveseldon))
- Ksonnet - initial integration [\#79](https://github.com/SeldonIO/seldon-core/pull/79) ([cliveseldon](https://github.com/cliveseldon))
- 54 epsilon greedy [\#78](https://github.com/SeldonIO/seldon-core/pull/78) ([Maximophone](https://github.com/Maximophone))

## [v0.1.3](https://github.com/SeldonIO/seldon-core/tree/v0.1.3) (2018-01-26)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.2...v0.1.3)

**Fixed bugs:**

- Insufficient cpu error when creating complex graphs [\#47](https://github.com/SeldonIO/seldon-core/issues/47)

**Closed issues:**

- Split Prometheus monitoring from seldon-core Helm Chart [\#69](https://github.com/SeldonIO/seldon-core/issues/69)
- Docs for how to deploy and CI/CD options [\#30](https://github.com/SeldonIO/seldon-core/issues/30)

**Merged pull requests:**

- Split Helm scripts into 3 - core, analytics and kafka [\#75](https://github.com/SeldonIO/seldon-core/pull/75) ([cliveseldon](https://github.com/cliveseldon))
- add engine resources to proto and modify cluster manager [\#72](https://github.com/SeldonIO/seldon-core/pull/72) ([cliveseldon](https://github.com/cliveseldon))
- 57 wrapping auto docs [\#68](https://github.com/SeldonIO/seldon-core/pull/68) ([Maximophone](https://github.com/Maximophone))
- Release v0.1.2 prep [\#66](https://github.com/SeldonIO/seldon-core/pull/66) ([gsunner](https://github.com/gsunner))

## [v0.1.2](https://github.com/SeldonIO/seldon-core/tree/v0.1.2) (2018-01-23)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.1...v0.1.2)

**Closed issues:**

- Remove the cluster manager functionality that checks for "type" presence in predictive units graph [\#59](https://github.com/SeldonIO/seldon-core/issues/59)
- Change builds to use seldonio/core-builder:0.2 [\#48](https://github.com/SeldonIO/seldon-core/issues/48)
- Cluster manager stuck in an error loop after failed deployment [\#45](https://github.com/SeldonIO/seldon-core/issues/45)
- Quickstart in minikube has old resource spec for deployment [\#40](https://github.com/SeldonIO/seldon-core/issues/40)
- Bring seldon-core-examples into main project [\#33](https://github.com/SeldonIO/seldon-core/issues/33)
- Update travis integration for all builds [\#25](https://github.com/SeldonIO/seldon-core/issues/25)
- Update model wrapping docs and docker wrapping code [\#17](https://github.com/SeldonIO/seldon-core/issues/17)
- Links are broken on the following doc page [\#16](https://github.com/SeldonIO/seldon-core/issues/16)

**Merged pull requests:**

- add travis build status [\#64](https://github.com/SeldonIO/seldon-core/pull/64) ([gsunner](https://github.com/gsunner))
- updates into Release 0.1 [\#63](https://github.com/SeldonIO/seldon-core/pull/63) ([gsunner](https://github.com/gsunner))
- add current release branch to travis builds [\#62](https://github.com/SeldonIO/seldon-core/pull/62) ([gsunner](https://github.com/gsunner))
- Complex graphs [\#61](https://github.com/SeldonIO/seldon-core/pull/61) ([Maximophone](https://github.com/Maximophone))
- change validation to handle no method check [\#60](https://github.com/SeldonIO/seldon-core/pull/60) ([cliveseldon](https://github.com/cliveseldon))
- update Quantity processing to allow non strings and catch exceptions in parsing protos [\#53](https://github.com/SeldonIO/seldon-core/pull/53) ([cliveseldon](https://github.com/cliveseldon))
- Dockerize the entier wrapping process of building sklearn\_iris example [\#51](https://github.com/SeldonIO/seldon-core/pull/51) ([errordeveloper](https://github.com/errordeveloper))
- travis builds updated to use core-builder:0.2 [\#49](https://github.com/SeldonIO/seldon-core/pull/49) ([gsunner](https://github.com/gsunner))
- use core-builder container for release script [\#46](https://github.com/SeldonIO/seldon-core/pull/46) ([gsunner](https://github.com/gsunner))
- add dependencies for the release script [\#44](https://github.com/SeldonIO/seldon-core/pull/44) ([gsunner](https://github.com/gsunner))
- Fixed json deployment [\#42](https://github.com/SeldonIO/seldon-core/pull/42) ([cliveseldon](https://github.com/cliveseldon))
- Updating minikube get started for newest version of the wrappers [\#41](https://github.com/SeldonIO/seldon-core/pull/41) ([Maximophone](https://github.com/Maximophone))
- helm yaml files updated for release script usage [\#39](https://github.com/SeldonIO/seldon-core/pull/39) ([gsunner](https://github.com/gsunner))
- release script code [\#38](https://github.com/SeldonIO/seldon-core/pull/38) ([gsunner](https://github.com/gsunner))
- 17 wrappers docs [\#37](https://github.com/SeldonIO/seldon-core/pull/37) ([Maximophone](https://github.com/Maximophone))
- Update to python wrapping: put the build and push docker image comman… [\#36](https://github.com/SeldonIO/seldon-core/pull/36) ([Maximophone](https://github.com/Maximophone))
- seldon-core-examples repo added to main project [\#34](https://github.com/SeldonIO/seldon-core/pull/34) ([gsunner](https://github.com/gsunner))
- add CI/CD docs [\#32](https://github.com/SeldonIO/seldon-core/pull/32) ([cliveseldon](https://github.com/cliveseldon))
- Travis update [\#31](https://github.com/SeldonIO/seldon-core/pull/31) ([gsunner](https://github.com/gsunner))
- Update docs crd [\#29](https://github.com/SeldonIO/seldon-core/pull/29) ([cliveseldon](https://github.com/cliveseldon))
- 17 wrappers docs [\#28](https://github.com/SeldonIO/seldon-core/pull/28) ([Maximophone](https://github.com/Maximophone))
- 17 wrappers update [\#27](https://github.com/SeldonIO/seldon-core/pull/27) ([Maximophone](https://github.com/Maximophone))

## [v0.1.1](https://github.com/SeldonIO/seldon-core/tree/v0.1.1) (2018-01-10)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.0...v0.1.1)

**Closed issues:**

- GRPC microservices fail kubernetes health and readiness checks [\#13](https://github.com/SeldonIO/seldon-core/issues/13)
- Update diagrams in docs to reflect latest version of the protos [\#7](https://github.com/SeldonIO/seldon-core/issues/7)
- Average Combiner broken in container [\#5](https://github.com/SeldonIO/seldon-core/issues/5)
- external gRPC client not updating when deployment removed and recreated [\#4](https://github.com/SeldonIO/seldon-core/issues/4)
- Cluster Manager failing to authenticate on GKE cluster [\#2](https://github.com/SeldonIO/seldon-core/issues/2)

**Merged pull requests:**

- Release 0.1 branch merge [\#22](https://github.com/SeldonIO/seldon-core/pull/22) ([gsunner](https://github.com/gsunner))
- version 0.1.1 prep [\#21](https://github.com/SeldonIO/seldon-core/pull/21) ([gsunner](https://github.com/gsunner))
- Health checks grpc [\#19](https://github.com/SeldonIO/seldon-core/pull/19) ([cliveseldon](https://github.com/cliveseldon))
- ci updates [\#18](https://github.com/SeldonIO/seldon-core/pull/18) ([gsunner](https://github.com/gsunner))
- 5 fix average combiner [\#14](https://github.com/SeldonIO/seldon-core/pull/14) ([Maximophone](https://github.com/Maximophone))
- Create CODE\_OF\_CONDUCT.md [\#12](https://github.com/SeldonIO/seldon-core/pull/12) ([cliveseldon](https://github.com/cliveseldon))
- Grpc apife publish [\#8](https://github.com/SeldonIO/seldon-core/pull/8) ([cliveseldon](https://github.com/cliveseldon))
- add RBAC serviceaccount [\#3](https://github.com/SeldonIO/seldon-core/pull/3) ([cliveseldon](https://github.com/cliveseldon))

## [v0.1.0](https://github.com/SeldonIO/seldon-core/tree/v0.1.0) (2018-01-03)


\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*