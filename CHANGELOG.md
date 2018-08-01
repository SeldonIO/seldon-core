# Change Log

## [v0.2.2](https://github.com/SeldonIO/seldon-core/tree/v0.2.2) (2018-07-18)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.1...v0.2.2)

**Merged pull requests:**

- Ksonnet update [\#191](https://github.com/SeldonIO/seldon-core/pull/191) ([cliveseldon](https://github.com/cliveseldon))
- Remove ambassador role and rolebinding from helm script [\#190](https://github.com/SeldonIO/seldon-core/pull/190) ([cliveseldon](https://github.com/cliveseldon))
- Update istio example [\#189](https://github.com/SeldonIO/seldon-core/pull/189) ([cliveseldon](https://github.com/cliveseldon))
- Update maven goals to generate licences [\#188](https://github.com/SeldonIO/seldon-core/pull/188) ([cliveseldon](https://github.com/cliveseldon))
- Fix typo [\#187](https://github.com/SeldonIO/seldon-core/pull/187) ([otakuto](https://github.com/otakuto))
- Istio updates [\#186](https://github.com/SeldonIO/seldon-core/pull/186) ([cliveseldon](https://github.com/cliveseldon))

## [v0.2.1](https://github.com/SeldonIO/seldon-core/tree/v0.2.1) (2018-07-09)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.0...v0.2.1)

**Closed issues:**

- "/s2i/bin/assemble: line 59:  10 Killed"  when using S2I to build PyTorch container  [\#180](https://github.com/SeldonIO/seldon-core/issues/180)
- Getting timeout error using S2I to package PyTorch model  [\#179](https://github.com/SeldonIO/seldon-core/issues/179)
- Make Operator Create CRD on StartUp [\#174](https://github.com/SeldonIO/seldon-core/issues/174)

**Merged pull requests:**

- Licences generation in poms [\#185](https://github.com/SeldonIO/seldon-core/pull/185) ([cliveseldon](https://github.com/cliveseldon))
- Update api testing utils to allow shape parameter [\#184](https://github.com/SeldonIO/seldon-core/pull/184) ([cliveseldon](https://github.com/cliveseldon))
- Ensure ambassador names are unique in resources created [\#182](https://github.com/SeldonIO/seldon-core/pull/182) ([cliveseldon](https://github.com/cliveseldon))
- Remove Application from helm chart [\#181](https://github.com/SeldonIO/seldon-core/pull/181) ([cliveseldon](https://github.com/cliveseldon))
- Updated Helm Chart and auto create of CRD [\#178](https://github.com/SeldonIO/seldon-core/pull/178) ([cliveseldon](https://github.com/cliveseldon))

## [v0.2.0](https://github.com/SeldonIO/seldon-core/tree/v0.2.0) (2018-06-29)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.8...v0.2.0)

**Closed issues:**

- PREDICTIVE\_UNIT\_PARAMETERS: not able to set them up correctly [\#170](https://github.com/SeldonIO/seldon-core/issues/170)
- Add docs page for Helm [\#71](https://github.com/SeldonIO/seldon-core/issues/71)

**Merged pull requests:**

- fix crd.libsonnet error [\#177](https://github.com/SeldonIO/seldon-core/pull/177) ([fisache](https://github.com/fisache))
- Distributed deployment and Istio [\#173](https://github.com/SeldonIO/seldon-core/pull/173) ([cliveseldon](https://github.com/cliveseldon))

## [v0.1.8](https://github.com/SeldonIO/seldon-core/tree/v0.1.8) (2018-06-27)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.7...v0.1.8)

**Closed issues:**

- Prediction analytics dashboard not capturing prediction API calls [\#168](https://github.com/SeldonIO/seldon-core/issues/168)
- Ambassador + Minikube doc needs to be updated per RBAC [\#165](https://github.com/SeldonIO/seldon-core/issues/165)
- configmap type volume gets mounted as EmptyDir [\#162](https://github.com/SeldonIO/seldon-core/issues/162)
- Java Wrapper H2OUtils doesn't check type in NDArray proto message conversion [\#158](https://github.com/SeldonIO/seldon-core/issues/158)
- Release Java wrappers library 0.1.1 [\#157](https://github.com/SeldonIO/seldon-core/issues/157)
- Automate update of ksonnet versions in release process [\#132](https://github.com/SeldonIO/seldon-core/issues/132)

**Merged pull requests:**

- Remove java wrapper library from code base [\#172](https://github.com/SeldonIO/seldon-core/pull/172) ([cliveseldon](https://github.com/cliveseldon))
- Remove nd4j and replace with oj matrix library [\#171](https://github.com/SeldonIO/seldon-core/pull/171) ([cliveseldon](https://github.com/cliveseldon))
- Fix prometheus helm install [\#169](https://github.com/SeldonIO/seldon-core/pull/169) ([cliveseldon](https://github.com/cliveseldon))
- Update notebooks for minikube and ambassador [\#166](https://github.com/SeldonIO/seldon-core/pull/166) ([cliveseldon](https://github.com/cliveseldon))
- Release script python3 compatibility [\#163](https://github.com/SeldonIO/seldon-core/pull/163) ([gsunner](https://github.com/gsunner))
- Updates to 0.1.1 wrapper. H2O fixes. [\#161](https://github.com/SeldonIO/seldon-core/pull/161) ([cliveseldon](https://github.com/cliveseldon))
- add update to core.jsonnet when setting version [\#159](https://github.com/SeldonIO/seldon-core/pull/159) ([gsunner](https://github.com/gsunner))

## [v0.1.7](https://github.com/SeldonIO/seldon-core/tree/v0.1.7) (2018-06-04)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.1.6...v0.1.7)

**Closed issues:**

- Quickstart problem [\#153](https://github.com/SeldonIO/seldon-core/issues/153)
- NameError: global name 'ListValue' is not defined [\#148](https://github.com/SeldonIO/seldon-core/issues/148)
- bad credentials error with get\_token function [\#144](https://github.com/SeldonIO/seldon-core/issues/144)
- Make CRD Namespaced scoped [\#141](https://github.com/SeldonIO/seldon-core/issues/141)
- Create wrappers for Java based models [\#137](https://github.com/SeldonIO/seldon-core/issues/137)
- Update ksonnet prototypes for latest image version [\#130](https://github.com/SeldonIO/seldon-core/issues/130)
- Create demo notebook for Azure [\#129](https://github.com/SeldonIO/seldon-core/issues/129)
- Grafana Dashboard  [\#109](https://github.com/SeldonIO/seldon-core/issues/109)
- Multiple helm seldon-core installs on separate namespaces fails [\#106](https://github.com/SeldonIO/seldon-core/issues/106)

**Merged pull requests:**

- Add install guide [\#156](https://github.com/SeldonIO/seldon-core/pull/156) ([cliveseldon](https://github.com/cliveseldon))
- WIP : PySpark and PMML example [\#155](https://github.com/SeldonIO/seldon-core/pull/155) ([cliveseldon](https://github.com/cliveseldon))
- Fix gRPC tests for wrappers and update sklearn iris example to show use [\#150](https://github.com/SeldonIO/seldon-core/pull/150) ([cliveseldon](https://github.com/cliveseldon))
- Minikube RBAC updates and Notebooks for Model examples [\#147](https://github.com/SeldonIO/seldon-core/pull/147) ([cliveseldon](https://github.com/cliveseldon))
- change ClusterRoleBinding to RoleBinding [\#146](https://github.com/SeldonIO/seldon-core/pull/146) ([gsunner](https://github.com/gsunner))
- MNIST loadtest [\#143](https://github.com/SeldonIO/seldon-core/pull/143) ([cliveseldon](https://github.com/cliveseldon))
- Openshift article on using s2i in seldon-core [\#140](https://github.com/SeldonIO/seldon-core/pull/140) ([cliveseldon](https://github.com/cliveseldon))
- Java wrappers [\#138](https://github.com/SeldonIO/seldon-core/pull/138) ([cliveseldon](https://github.com/cliveseldon))
- add notebook for azure demo [\#135](https://github.com/SeldonIO/seldon-core/pull/135) ([gsunner](https://github.com/gsunner))
- update ksonnet defaults to 0.1.6 [\#131](https://github.com/SeldonIO/seldon-core/pull/131) ([cliveseldon](https://github.com/cliveseldon))
- Typos fix [\#128](https://github.com/SeldonIO/seldon-core/pull/128) ([LevineHuang](https://github.com/LevineHuang))

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