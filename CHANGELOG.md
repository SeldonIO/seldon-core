# Change Log

## [v0.2.5](https://github.com/SeldonIO/seldon-core/tree/v0.2.5) (2018-12-16)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.4...v0.2.5)

**Closed issues:**

- initialDelaySeconds: 10 sec is not enough for some models [\#323](https://github.com/SeldonIO/seldon-core/issues/323)
- Bug: custom metrics for both children and parent components [\#322](https://github.com/SeldonIO/seldon-core/issues/322)
- Ambassador seldon deployment not registered [\#318](https://github.com/SeldonIO/seldon-core/issues/318)
- Allow user to disable Redis in seldon-core helm chart [\#304](https://github.com/SeldonIO/seldon-core/issues/304)
- grpc tensor convert not valid for python 2 [\#301](https://github.com/SeldonIO/seldon-core/issues/301)
- Ambassador  [\#298](https://github.com/SeldonIO/seldon-core/issues/298)
- Create a python wrapper for COMBINER components [\#296](https://github.com/SeldonIO/seldon-core/issues/296)
- packaging for python microservice wrapper [\#293](https://github.com/SeldonIO/seldon-core/issues/293)
- Update to latest Spartakus image [\#291](https://github.com/SeldonIO/seldon-core/issues/291)
- Docker image "seldonio/seldon-core-s2i-python3" uses old Python 3.6 [\#288](https://github.com/SeldonIO/seldon-core/issues/288)
- Seldon cluster-manager  k8s cluster wide operations [\#269](https://github.com/SeldonIO/seldon-core/issues/269)
- S2I hangs in example models when using minikube docker-env [\#253](https://github.com/SeldonIO/seldon-core/issues/253)
- Accessing custom metrics in our Python model [\#245](https://github.com/SeldonIO/seldon-core/issues/245)
- strData & binData not accepted by Python model microservice [\#225](https://github.com/SeldonIO/seldon-core/issues/225)
- Utilize latest /status endpoint for Custom Resources in k8s 1.11 [\#176](https://github.com/SeldonIO/seldon-core/issues/176)
- Investigate Nvidia's TensorRT [\#121](https://github.com/SeldonIO/seldon-core/issues/121)
- Review status field for CRD [\#83](https://github.com/SeldonIO/seldon-core/issues/83)
- gitops demo [\#67](https://github.com/SeldonIO/seldon-core/issues/67)
- Update seldon-core/examples docs after move [\#35](https://github.com/SeldonIO/seldon-core/issues/35)
- Create wrapper for Spark standalone runtime models [\#24](https://github.com/SeldonIO/seldon-core/issues/24)

**Merged pull requests:**

- Update e2e tests and add Combiner to python wrappers [\#343](https://github.com/SeldonIO/seldon-core/pull/343) ([cliveseldon](https://github.com/cliveseldon))
- Python wrapper update and openvino example [\#342](https://github.com/SeldonIO/seldon-core/pull/342) ([cliveseldon](https://github.com/cliveseldon))
- Remove legacy python wrapper modules [\#339](https://github.com/SeldonIO/seldon-core/pull/339) ([jklaise](https://github.com/jklaise))
- Update S2I version in examples [\#338](https://github.com/SeldonIO/seldon-core/pull/338) ([cliveseldon](https://github.com/cliveseldon))
- Cluster Wide Operator [\#334](https://github.com/SeldonIO/seldon-core/pull/334) ([cliveseldon](https://github.com/cliveseldon))
- update python requests package version [\#331](https://github.com/SeldonIO/seldon-core/pull/331) ([arnaudvl](https://github.com/arnaudvl))
- Fix bug in parent custom metrics [\#329](https://github.com/SeldonIO/seldon-core/pull/329) ([cliveseldon](https://github.com/cliveseldon))
- ResNet Latency test [\#328](https://github.com/SeldonIO/seldon-core/pull/328) ([cliveseldon](https://github.com/cliveseldon))
- adding isolation forest and reorganize vae [\#327](https://github.com/SeldonIO/seldon-core/pull/327) ([arnaudvl](https://github.com/arnaudvl))
- Add serving doc [\#326](https://github.com/SeldonIO/seldon-core/pull/326) ([cliveseldon](https://github.com/cliveseldon))
- Update docs for API examples and latest protos [\#325](https://github.com/SeldonIO/seldon-core/pull/325) ([cliveseldon](https://github.com/cliveseldon))
- Allow further options for binary and tensors in prediction API [\#321](https://github.com/SeldonIO/seldon-core/pull/321) ([cliveseldon](https://github.com/cliveseldon))
- outlier detection component [\#320](https://github.com/SeldonIO/seldon-core/pull/320) ([arnaudvl](https://github.com/arnaudvl))
- Fix grpc tensor convert for python2 [\#317](https://github.com/SeldonIO/seldon-core/pull/317) ([cliveseldon](https://github.com/cliveseldon))
- Fix bug in parsing truth values for feedback [\#316](https://github.com/SeldonIO/seldon-core/pull/316) ([jklaise](https://github.com/jklaise))
- WIP: Custom metric tags [\#311](https://github.com/SeldonIO/seldon-core/pull/311) ([cliveseldon](https://github.com/cliveseldon))
- Fix tester docs to point to correct links [\#307](https://github.com/SeldonIO/seldon-core/pull/307) ([jklaise](https://github.com/jklaise))
- Create initial Python package [\#306](https://github.com/SeldonIO/seldon-core/pull/306) ([jklaise](https://github.com/jklaise))
- Allow disable redis [\#305](https://github.com/SeldonIO/seldon-core/pull/305) ([ChenyuanZ](https://github.com/ChenyuanZ))
- fix status remove functionality in operator [\#300](https://github.com/SeldonIO/seldon-core/pull/300) ([cliveseldon](https://github.com/cliveseldon))
- Fix storing of Gauge metrics [\#299](https://github.com/SeldonIO/seldon-core/pull/299) ([cliveseldon](https://github.com/cliveseldon))
- Rolling Update Fixes [\#295](https://github.com/SeldonIO/seldon-core/pull/295) ([cliveseldon](https://github.com/cliveseldon))
- Update java base images [\#292](https://github.com/SeldonIO/seldon-core/pull/292) ([cliveseldon](https://github.com/cliveseldon))
- WIP: Create python 3.6 and 3.7 wrapper versions [\#290](https://github.com/SeldonIO/seldon-core/pull/290) ([cliveseldon](https://github.com/cliveseldon))
- Custom Metrics [\#281](https://github.com/SeldonIO/seldon-core/pull/281) ([cliveseldon](https://github.com/cliveseldon))

## [v0.2.4](https://github.com/SeldonIO/seldon-core/tree/v0.2.4) (2018-11-07)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.3...v0.2.4)

**Closed issues:**

- Specification of a Service Account [\#286](https://github.com/SeldonIO/seldon-core/issues/286)
- curl not found error [\#283](https://github.com/SeldonIO/seldon-core/issues/283)
- Allow ambassador from other namespace to access SeldonDeployment [\#279](https://github.com/SeldonIO/seldon-core/issues/279)
- Fix Github security vulnerabilities in dependencies [\#259](https://github.com/SeldonIO/seldon-core/issues/259)
- Feedback API not called when using Models [\#251](https://github.com/SeldonIO/seldon-core/issues/251)
- Allow JAVA OPTS for engine to be specified [\#249](https://github.com/SeldonIO/seldon-core/issues/249)
- ndarray greater than 15280 bytes [\#248](https://github.com/SeldonIO/seldon-core/issues/248)
- Prediction API get model version [\#244](https://github.com/SeldonIO/seldon-core/issues/244)
- SeldonDeployment creation strips out an empty "children" list field in the manifest [\#242](https://github.com/SeldonIO/seldon-core/issues/242)
- Mahalanobis Outlier Detector fails when batch is of size 1 [\#240](https://github.com/SeldonIO/seldon-core/issues/240)
- Seldon Core Operator defaulting causes issues with helm and ArgoCD [\#233](https://github.com/SeldonIO/seldon-core/issues/233)
- TensorFlow Serving as the Model microservice [\#226](https://github.com/SeldonIO/seldon-core/issues/226)
- Nodejs wrapper for javascript models [\#216](https://github.com/SeldonIO/seldon-core/issues/216)
- Environmental variable error [\#215](https://github.com/SeldonIO/seldon-core/issues/215)
- How do I increase timeout of sidecar seldon container? [\#196](https://github.com/SeldonIO/seldon-core/issues/196)
- update release script for pyhton3 [\#160](https://github.com/SeldonIO/seldon-core/issues/160)
- Ability to customize Ambassador configuration [\#120](https://github.com/SeldonIO/seldon-core/issues/120)
- Script to convert proto files and generate OpenAPI schema [\#9](https://github.com/SeldonIO/seldon-core/issues/9)
- OpenAPI spec for external and internal prediction APIs [\#6](https://github.com/SeldonIO/seldon-core/issues/6)

**Merged pull requests:**

- Add optional service account for engine [\#287](https://github.com/SeldonIO/seldon-core/pull/287) ([cliveseldon](https://github.com/cliveseldon))
- Add missing curl to engine Dockerfile [\#285](https://github.com/SeldonIO/seldon-core/pull/285) ([cliveseldon](https://github.com/cliveseldon))
- Allow ambassador from other namespace to access SeldonDeployment [\#280](https://github.com/SeldonIO/seldon-core/pull/280) ([ChenyuanZ](https://github.com/ChenyuanZ))
- Faster protobuffer to numpy conversion in python wrapper [\#278](https://github.com/SeldonIO/seldon-core/pull/278) ([cliveseldon](https://github.com/cliveseldon))
- Ensure cluster role has unique name  [\#276](https://github.com/SeldonIO/seldon-core/pull/276) ([cliveseldon](https://github.com/cliveseldon))
- fix api-tester not using oauth-key and oauth-secret args [\#275](https://github.com/SeldonIO/seldon-core/pull/275) ([gsunner](https://github.com/gsunner))
- Update when status is set [\#273](https://github.com/SeldonIO/seldon-core/pull/273) ([cliveseldon](https://github.com/cliveseldon))
- Add OUTPUT\_TRANSFORMER example [\#272](https://github.com/SeldonIO/seldon-core/pull/272) ([ChenyuanZ](https://github.com/ChenyuanZ))
- Add Open API Definitions [\#271](https://github.com/SeldonIO/seldon-core/pull/271) ([cliveseldon](https://github.com/cliveseldon))
- Update Custom Resources via k8s /status endpoint if possible [\#270](https://github.com/SeldonIO/seldon-core/pull/270) ([cliveseldon](https://github.com/cliveseldon))
- Fixed small copy-paste error [\#266](https://github.com/SeldonIO/seldon-core/pull/266) ([lorello](https://github.com/lorello))
- Update ambassador to 0.40.0 [\#265](https://github.com/SeldonIO/seldon-core/pull/265) ([cliveseldon](https://github.com/cliveseldon))
- Add code coverage Jacoco to poms [\#264](https://github.com/SeldonIO/seldon-core/pull/264) ([cliveseldon](https://github.com/cliveseldon))
- Fix vulnerability warnings with updates to engine and apife pom [\#263](https://github.com/SeldonIO/seldon-core/pull/263) ([cliveseldon](https://github.com/cliveseldon))
- Add custom metrics proposal [\#261](https://github.com/SeldonIO/seldon-core/pull/261) ([cliveseldon](https://github.com/cliveseldon))
- Intel Openvino Integration [\#260](https://github.com/SeldonIO/seldon-core/pull/260) ([cliveseldon](https://github.com/cliveseldon))
- Python wrapper custom endpoints [\#257](https://github.com/SeldonIO/seldon-core/pull/257) ([gsunner](https://github.com/gsunner))
- Sending Feedback to Models [\#254](https://github.com/SeldonIO/seldon-core/pull/254) ([cliveseldon](https://github.com/cliveseldon))
- Python wrapper custom endpoints [\#252](https://github.com/SeldonIO/seldon-core/pull/252) ([gsunner](https://github.com/gsunner))
- Engine java opts annotations and ambassador timeout annotation [\#250](https://github.com/SeldonIO/seldon-core/pull/250) ([cliveseldon](https://github.com/cliveseldon))
- Update Ksonnet and Helm Charts [\#247](https://github.com/SeldonIO/seldon-core/pull/247) ([cliveseldon](https://github.com/cliveseldon))
- Add requestPath to response meta data [\#246](https://github.com/SeldonIO/seldon-core/pull/246) ([cliveseldon](https://github.com/cliveseldon))
- Fix outlier detection divide by zero and add initial mnist example \(wip\) [\#243](https://github.com/SeldonIO/seldon-core/pull/243) ([cliveseldon](https://github.com/cliveseldon))
- Fix typos in docs [\#241](https://github.com/SeldonIO/seldon-core/pull/241) ([jklaise](https://github.com/jklaise))
- Add example helm charts for inference graphs [\#239](https://github.com/SeldonIO/seldon-core/pull/239) ([cliveseldon](https://github.com/cliveseldon))
- Fix for defaulting changing Custom Resource [\#238](https://github.com/SeldonIO/seldon-core/pull/238) ([cliveseldon](https://github.com/cliveseldon))
- Image pull policy ksonnet fix [\#235](https://github.com/SeldonIO/seldon-core/pull/235) ([gsunner](https://github.com/gsunner))
- Nvidia Inference Server and Tensorflow Serving Model Proxies [\#234](https://github.com/SeldonIO/seldon-core/pull/234) ([cliveseldon](https://github.com/cliveseldon))
- Update kubectl\_demo\_minikube\_rbac.ipynb [\#232](https://github.com/SeldonIO/seldon-core/pull/232) ([benoitbayol](https://github.com/benoitbayol))
- Update epsilon-greedy example to Python 3 [\#231](https://github.com/SeldonIO/seldon-core/pull/231) ([jklaise](https://github.com/jklaise))
- Update kubectl\_demo\_minikube\_rbac.ipynb [\#230](https://github.com/SeldonIO/seldon-core/pull/230) ([benoitbayol](https://github.com/benoitbayol))
- GRPC API for javascript models with Nodejs s2i wrapper [\#224](https://github.com/SeldonIO/seldon-core/pull/224) ([SachinVarghese](https://github.com/SachinVarghese))

## [v0.2.3](https://github.com/SeldonIO/seldon-core/tree/v0.2.3) (2018-09-17)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.2...v0.2.3)

**Closed issues:**

- R wrapper s2i environment documentation missing Model file extension [\#219](https://github.com/SeldonIO/seldon-core/issues/219)
- Provide example using ONNX via Intel nGraph for inference [\#214](https://github.com/SeldonIO/seldon-core/issues/214)
- how to explore grafana dashboard for seldon-core  in  web ui [\#209](https://github.com/SeldonIO/seldon-core/issues/209)
- APPLICATION FAILED TO START - Example python notebook \(fx-market-predictor\) [\#208](https://github.com/SeldonIO/seldon-core/issues/208)
- dev guide doc: develop/test changes locally? [\#202](https://github.com/SeldonIO/seldon-core/issues/202)
- Service orchestrator updated when surrounding deployment changed [\#199](https://github.com/SeldonIO/seldon-core/issues/199)
- UnknownHostException: seldon-deployment [\#194](https://github.com/SeldonIO/seldon-core/issues/194)
- grafana dashboard [\#192](https://github.com/SeldonIO/seldon-core/issues/192)
- Add image versions to all wrapper images [\#136](https://github.com/SeldonIO/seldon-core/issues/136)
- Status is not created in Custom Resource on initial create or update [\#74](https://github.com/SeldonIO/seldon-core/issues/74)

**Merged pull requests:**

- Update SeldonDeployment status for lifecycle conditions [\#223](https://github.com/SeldonIO/seldon-core/pull/223) ([cliveseldon](https://github.com/cliveseldon))
- Update use of python wrappers to version 0.2 [\#222](https://github.com/SeldonIO/seldon-core/pull/222) ([cliveseldon](https://github.com/cliveseldon))
- Support for ONNX exported models for inference [\#221](https://github.com/SeldonIO/seldon-core/pull/221) ([cliveseldon](https://github.com/cliveseldon))
- Private repo build and run [\#220](https://github.com/SeldonIO/seldon-core/pull/220) ([gsunner](https://github.com/gsunner))
- Nodejs s2i wrapper for JavaScript models [\#218](https://github.com/SeldonIO/seldon-core/pull/218) ([SachinVarghese](https://github.com/SachinVarghese))
- Update seldon metrics [\#217](https://github.com/SeldonIO/seldon-core/pull/217) ([cliveseldon](https://github.com/cliveseldon))
- fix oauth\_port check in api-tester [\#213](https://github.com/SeldonIO/seldon-core/pull/213) ([cliveseldon](https://github.com/cliveseldon))
- Add configurable timeouts for REST and gRPC [\#211](https://github.com/SeldonIO/seldon-core/pull/211) ([cliveseldon](https://github.com/cliveseldon))
- Config circular bug [\#210](https://github.com/SeldonIO/seldon-core/pull/210) ([cliveseldon](https://github.com/cliveseldon))
- Update for flatbuffers python wrappers [\#205](https://github.com/SeldonIO/seldon-core/pull/205) ([cliveseldon](https://github.com/cliveseldon))
- Experimental Flatbuffers based protocol for python wrapper [\#204](https://github.com/SeldonIO/seldon-core/pull/204) ([cliveseldon](https://github.com/cliveseldon))
- Change wrappers to be versioned and update examples and docs [\#201](https://github.com/SeldonIO/seldon-core/pull/201) ([cliveseldon](https://github.com/cliveseldon))
- Update istio example notebook [\#200](https://github.com/SeldonIO/seldon-core/pull/200) ([cliveseldon](https://github.com/cliveseldon))
- Allow Annotations to allow customizations [\#197](https://github.com/SeldonIO/seldon-core/pull/197) ([cliveseldon](https://github.com/cliveseldon))
- Removed subtype from deployment example [\#195](https://github.com/SeldonIO/seldon-core/pull/195) ([hanneshapke](https://github.com/hanneshapke))

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