# Change Log

## [v0.3.0](https://github.com/SeldonIO/seldon-core/tree/v0.3.0) (2019-06-05)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.7...v0.3.0)

**Fixed bugs:**

- Python module syntax error for python 3.5 for variable type annotation [\#614](https://github.com/SeldonIO/seldon-core/issues/614)
- openvino\_imagenet\_ensemble example - prediction1 and prediction2 containers error [\#583](https://github.com/SeldonIO/seldon-core/issues/583)
- Docker image name and version messed up [\#562](https://github.com/SeldonIO/seldon-core/issues/562)
- SendonDeployment with name \> 31 characters [\#556](https://github.com/SeldonIO/seldon-core/issues/556)
- Not able be build python3.6 base image.  [\#542](https://github.com/SeldonIO/seldon-core/issues/542)

**Closed issues:**

- Update master to 0.3.0 SNAPSHOT [\#612](https://github.com/SeldonIO/seldon-core/issues/612)
- sdep state doesn't move to available [\#605](https://github.com/SeldonIO/seldon-core/issues/605)
- engine using 1 cpu [\#597](https://github.com/SeldonIO/seldon-core/issues/597)
- forbidden error installing ambassador [\#596](https://github.com/SeldonIO/seldon-core/issues/596)
- GPU support with SERVICE\_TYPE Model [\#590](https://github.com/SeldonIO/seldon-core/issues/590)
- Update example notebooks for docs [\#586](https://github.com/SeldonIO/seldon-core/issues/586)
- kubeflow/example-seldon on local cluster [\#585](https://github.com/SeldonIO/seldon-core/issues/585)
- Unable to inject custom parameter in Python model [\#584](https://github.com/SeldonIO/seldon-core/issues/584)
- Tensorflow MNIST Model example on EKS [\#580](https://github.com/SeldonIO/seldon-core/issues/580)
- OOMKilled when starting an operator [\#579](https://github.com/SeldonIO/seldon-core/issues/579)
- Can we customize the outputs format of Model? [\#565](https://github.com/SeldonIO/seldon-core/issues/565)
- node exporter port conflict [\#563](https://github.com/SeldonIO/seldon-core/issues/563)
- How can i enable debug logging of seldon-engine container [\#560](https://github.com/SeldonIO/seldon-core/issues/560)
- Tensorflow Python 3.7 support and wrapper images [\#550](https://github.com/SeldonIO/seldon-core/issues/550)
- helm-charts upgrade fails on ambassador [\#543](https://github.com/SeldonIO/seldon-core/issues/543)
- Update Seldon Core Analytics Grafana [\#540](https://github.com/SeldonIO/seldon-core/issues/540)
- Defining the model serving class with full name doesn't currently work [\#533](https://github.com/SeldonIO/seldon-core/issues/533)
- Update Seldon Operator to Go [\#529](https://github.com/SeldonIO/seldon-core/issues/529)
- Old Containers & Security Vulnerabilities [\#528](https://github.com/SeldonIO/seldon-core/issues/528)
- option to not set runAsUser for engine [\#527](https://github.com/SeldonIO/seldon-core/issues/527)
- Support istio ingress [\#511](https://github.com/SeldonIO/seldon-core/issues/511)
- Endpoint type is missing for AB-test [\#451](https://github.com/SeldonIO/seldon-core/issues/451)
- Alllow arbitrary JSON as a payload [\#434](https://github.com/SeldonIO/seldon-core/issues/434)
- Update Ambassador to 0.40.2 [\#402](https://github.com/SeldonIO/seldon-core/issues/402)
- Ambassador config for rolling updates [\#294](https://github.com/SeldonIO/seldon-core/issues/294)

**Merged pull requests:**

- node exporter configurable port [\#617](https://github.com/SeldonIO/seldon-core/pull/617) ([csabika7](https://github.com/csabika7))
- Require python 3.6 or above for python module [\#615](https://github.com/SeldonIO/seldon-core/pull/615) ([cliveseldon](https://github.com/cliveseldon))
- Update python wrappers to include 3.7 [\#611](https://github.com/SeldonIO/seldon-core/pull/611) ([cliveseldon](https://github.com/cliveseldon))
- update jackson libs to version 2.9.9 [\#609](https://github.com/SeldonIO/seldon-core/pull/609) ([gsunner](https://github.com/gsunner))
- ambassador v1 api [\#603](https://github.com/SeldonIO/seldon-core/pull/603) ([ryandawsonuk](https://github.com/ryandawsonuk))
- option to not set engine user [\#601](https://github.com/SeldonIO/seldon-core/pull/601) ([ryandawsonuk](https://github.com/ryandawsonuk))
- take latest ambassador image [\#599](https://github.com/SeldonIO/seldon-core/pull/599) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Update SeldonMessage with jsonData [\#595](https://github.com/SeldonIO/seldon-core/pull/595) ([gsunner](https://github.com/gsunner))
- Python release update [\#594](https://github.com/SeldonIO/seldon-core/pull/594) ([jklaise](https://github.com/jklaise))
- Fix Jupyter Notebook Headers  [\#592](https://github.com/SeldonIO/seldon-core/pull/592) ([axsauze](https://github.com/axsauze))
- Kubeflow Seldon e2e NLP ML pipeline using re-usable components [\#589](https://github.com/SeldonIO/seldon-core/pull/589) ([axsauze](https://github.com/axsauze))
- WIP: Integrate with Istio Ingress [\#588](https://github.com/SeldonIO/seldon-core/pull/588) ([cliveseldon](https://github.com/cliveseldon))
- Added missing link to Jupyter notebook [\#587](https://github.com/SeldonIO/seldon-core/pull/587) ([axsauze](https://github.com/axsauze))
- Added missed s2i folder to Scikitlearn SpaCy Text Example [\#582](https://github.com/SeldonIO/seldon-core/pull/582) ([axsauze](https://github.com/axsauze))
- AWS Elastic Kubernetes/Container Service Deep Mnist Example [\#581](https://github.com/SeldonIO/seldon-core/pull/581) ([axsauze](https://github.com/axsauze))
- Example using Seldon for text classification with SpaCy tokenizer [\#578](https://github.com/SeldonIO/seldon-core/pull/578) ([axsauze](https://github.com/axsauze))
- Remove request limits from operator [\#577](https://github.com/SeldonIO/seldon-core/pull/577) ([cliveseldon](https://github.com/cliveseldon))
- Fix PredictiveUnitState image name and version \(\#562\) [\#576](https://github.com/SeldonIO/seldon-core/pull/576) ([sasvaritoni](https://github.com/sasvaritoni))
- Update TF version for security [\#575](https://github.com/SeldonIO/seldon-core/pull/575) ([jklaise](https://github.com/jklaise))
- updated openvino mode ensemble to 0.2 version [\#574](https://github.com/SeldonIO/seldon-core/pull/574) ([dtrawins](https://github.com/dtrawins))
- updated openvino version to 2019.1 in python\_openvino model wrapper [\#573](https://github.com/SeldonIO/seldon-core/pull/573) ([dtrawins](https://github.com/dtrawins))
- Fix example deployment yaml [\#571](https://github.com/SeldonIO/seldon-core/pull/571) ([sujaymansingh](https://github.com/sujaymansingh))
- fix typo [\#570](https://github.com/SeldonIO/seldon-core/pull/570) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Update Python builder image [\#568](https://github.com/SeldonIO/seldon-core/pull/568) ([jklaise](https://github.com/jklaise))
- option for R builds with plain docker [\#567](https://github.com/SeldonIO/seldon-core/pull/567) ([ryandawsonuk](https://github.com/ryandawsonuk))
- reword explanation of ambassador [\#561](https://github.com/SeldonIO/seldon-core/pull/561) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Integrate use of Go Seldon Controller [\#559](https://github.com/SeldonIO/seldon-core/pull/559) ([cliveseldon](https://github.com/cliveseldon))
- Update e2e tests s2i python image version [\#558](https://github.com/SeldonIO/seldon-core/pull/558) ([gsunner](https://github.com/gsunner))
- option for docker build without s2i [\#555](https://github.com/SeldonIO/seldon-core/pull/555) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Fix logging bug in Python wrapper [\#549](https://github.com/SeldonIO/seldon-core/pull/549) ([jklaise](https://github.com/jklaise))
- update jackson-databind 2.8.11.2 -\> 2.9.8 for cve [\#547](https://github.com/SeldonIO/seldon-core/pull/547) ([gsunner](https://github.com/gsunner))
- Updating grafana to v6.1.6 in seldon core analytics [\#541](https://github.com/SeldonIO/seldon-core/pull/541) ([SachinVarghese](https://github.com/SachinVarghese))
- redis now a statefulset as using redis helm chart [\#539](https://github.com/SeldonIO/seldon-core/pull/539) ([ryandawsonuk](https://github.com/ryandawsonuk))
- add script to delete completed argo jobs [\#538](https://github.com/SeldonIO/seldon-core/pull/538) ([gsunner](https://github.com/gsunner))
- Allow fully qualified class name to be used for the model serving image. [\#537](https://github.com/SeldonIO/seldon-core/pull/537) ([hmonteiro](https://github.com/hmonteiro))
- option for anonymous access to grafana [\#535](https://github.com/SeldonIO/seldon-core/pull/535) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Anonymous grafana [\#534](https://github.com/SeldonIO/seldon-core/pull/534) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Revert "option to use anonymous auth grafana" [\#532](https://github.com/SeldonIO/seldon-core/pull/532) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Update component code coverage and dependencies docs [\#531](https://github.com/SeldonIO/seldon-core/pull/531) ([cliveseldon](https://github.com/cliveseldon))
- update argocd and jenkins in cd demo and script for minikube [\#517](https://github.com/SeldonIO/seldon-core/pull/517) ([ryandawsonuk](https://github.com/ryandawsonuk))

## [v0.2.7](https://github.com/SeldonIO/seldon-core/tree/v0.2.7) (2019-04-29)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.6...v0.2.7)

**Implemented enhancements:**

- Type check predictive unit parameters in the Python wrapper [\#440](https://github.com/SeldonIO/seldon-core/issues/440)

**Fixed bugs:**

- Models pods duplications after corrupted deployment [\#470](https://github.com/SeldonIO/seldon-core/issues/470)
- Using a configMapRef inside of a seldon deployment manifest causes a NullPointerException in the SeldonDeploymentWatcher [\#450](https://github.com/SeldonIO/seldon-core/issues/450)
- cannot get working external api but internal api is ok [\#448](https://github.com/SeldonIO/seldon-core/issues/448)
- Status can become Available even with Exception in Operator [\#429](https://github.com/SeldonIO/seldon-core/issues/429)
- Fix status update for failed deployments [\#474](https://github.com/SeldonIO/seldon-core/pull/474) ([cliveseldon](https://github.com/cliveseldon))

**Closed issues:**

- Install seldon in a single namespace with restricted tiller [\#514](https://github.com/SeldonIO/seldon-core/issues/514)
- Document about microservice's input data [\#512](https://github.com/SeldonIO/seldon-core/issues/512)
- where is io.seldon.protos.DeploymentProtos package located? [\#508](https://github.com/SeldonIO/seldon-core/issues/508)
- seldon 0.2.3 - nfs volume in seldon graph failing in validation [\#504](https://github.com/SeldonIO/seldon-core/issues/504)
- SeldonDeployment keeps hanging [\#499](https://github.com/SeldonIO/seldon-core/issues/499)
- default ambassador chart to single namespace [\#495](https://github.com/SeldonIO/seldon-core/issues/495)
- use v1 ambassador api [\#491](https://github.com/SeldonIO/seldon-core/issues/491)
- Configure the way Prometheus exposed [\#484](https://github.com/SeldonIO/seldon-core/issues/484)
- documentation is in doc not docs [\#481](https://github.com/SeldonIO/seldon-core/issues/481)
- do a snapshot build and document if not documented [\#479](https://github.com/SeldonIO/seldon-core/issues/479)
- How can we specify nested python class in .s2i/environment? [\#465](https://github.com/SeldonIO/seldon-core/issues/465)
- Class names in latest python library is not backwards compatible [\#462](https://github.com/SeldonIO/seldon-core/issues/462)
- Sending an object dtype array as the request JSON for a Model API [\#461](https://github.com/SeldonIO/seldon-core/issues/461)
- NullPointer exception in API gateway when principal can't be determined [\#454](https://github.com/SeldonIO/seldon-core/issues/454)
- Python Wrappers Version 2 [\#406](https://github.com/SeldonIO/seldon-core/issues/406)
- Write a Python wrapper for a GENERIC component [\#378](https://github.com/SeldonIO/seldon-core/issues/378)
- Create reference Python client [\#349](https://github.com/SeldonIO/seldon-core/issues/349)
- Python-wrapper: Use debug flag to provide useful information [\#309](https://github.com/SeldonIO/seldon-core/issues/309)
- Support autoscaler in SeldonDeployment [\#277](https://github.com/SeldonIO/seldon-core/issues/277)
- Update Ambassdor Helm or remove and use Ambassador's helm chart [\#258](https://github.com/SeldonIO/seldon-core/issues/258)
- Prow Integration [\#154](https://github.com/SeldonIO/seldon-core/issues/154)
- CI/CD demo using GitOps framework [\#11](https://github.com/SeldonIO/seldon-core/issues/11)

**Merged pull requests:**

- option to use anonymous auth grafana [\#530](https://github.com/SeldonIO/seldon-core/pull/530) ([ryandawsonuk](https://github.com/ryandawsonuk))
- permission and timeout changes after trying on an openshift4 cluster [\#524](https://github.com/SeldonIO/seldon-core/pull/524) ([ryandawsonuk](https://github.com/ryandawsonuk))
- use stable redis helm chart [\#521](https://github.com/SeldonIO/seldon-core/pull/521) ([ryandawsonuk](https://github.com/ryandawsonuk))
- seldpon\_grpc\_endpoint -\> seldon\_grpc\_endpoint [\#520](https://github.com/SeldonIO/seldon-core/pull/520) ([mustyoshi](https://github.com/mustyoshi))
- Service Orchestrator Name Fix [\#516](https://github.com/SeldonIO/seldon-core/pull/516) ([cliveseldon](https://github.com/cliveseldon))
- Remove v1alpha3 and revert to v1alpha2 [\#513](https://github.com/SeldonIO/seldon-core/pull/513) ([cliveseldon](https://github.com/cliveseldon))
- downgrade ambassador [\#510](https://github.com/SeldonIO/seldon-core/pull/510) ([ryandawsonuk](https://github.com/ryandawsonuk))
- default ambassador to singleNamespace [\#509](https://github.com/SeldonIO/seldon-core/pull/509) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Allow submodules to be imported in python module [\#503](https://github.com/SeldonIO/seldon-core/pull/503) ([cliveseldon](https://github.com/cliveseldon))
- Allow class\_names as method or attribute \(deprecated\) in Python module [\#502](https://github.com/SeldonIO/seldon-core/pull/502) ([cliveseldon](https://github.com/cliveseldon))
- downgrade ambassador due to grpc unreliability [\#501](https://github.com/SeldonIO/seldon-core/pull/501) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Fix HPA Nullpointer [\#500](https://github.com/SeldonIO/seldon-core/pull/500) ([cliveseldon](https://github.com/cliveseldon))
- still intermittent problems, timeout needs to be longer [\#498](https://github.com/SeldonIO/seldon-core/pull/498) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Add missing additionProperties to openAPI specs for CRDS [\#496](https://github.com/SeldonIO/seldon-core/pull/496) ([cliveseldon](https://github.com/cliveseldon))
- Spelling [\#493](https://github.com/SeldonIO/seldon-core/pull/493) ([mustyoshi](https://github.com/mustyoshi))
- ambassador v1 api [\#492](https://github.com/SeldonIO/seldon-core/pull/492) ([ryandawsonuk](https://github.com/ryandawsonuk))
- Fix image link in readme [\#490](https://github.com/SeldonIO/seldon-core/pull/490) ([cliveseldon](https://github.com/cliveseldon))
- Updates for various Python and Operator fixes [\#488](https://github.com/SeldonIO/seldon-core/pull/488) ([cliveseldon](https://github.com/cliveseldon))
- 484 metrics port [\#485](https://github.com/SeldonIO/seldon-core/pull/485) ([ryandawsonuk](https://github.com/ryandawsonuk))
- ignore pickle files [\#483](https://github.com/SeldonIO/seldon-core/pull/483) ([ryandawsonuk](https://github.com/ryandawsonuk))
- remove old docs [\#482](https://github.com/SeldonIO/seldon-core/pull/482) ([ryandawsonuk](https://github.com/ryandawsonuk))
- make ambassador a dependency [\#480](https://github.com/SeldonIO/seldon-core/pull/480) ([ryandawsonuk](https://github.com/ryandawsonuk))
- gitignore for intellij [\#471](https://github.com/SeldonIO/seldon-core/pull/471) ([ryandawsonuk](https://github.com/ryandawsonuk))
- python wrapper image fix update [\#469](https://github.com/SeldonIO/seldon-core/pull/469) ([gsunner](https://github.com/gsunner))
- python wrapper image references updated from 0.5 to 0.5.1 [\#468](https://github.com/SeldonIO/seldon-core/pull/468) ([gsunner](https://github.com/gsunner))
- Static Documentation Site [\#466](https://github.com/SeldonIO/seldon-core/pull/466) ([cliveseldon](https://github.com/cliveseldon))
- Remove tornando dependency from Python setup.py [\#464](https://github.com/SeldonIO/seldon-core/pull/464) ([cliveseldon](https://github.com/cliveseldon))
- Add types for predict, transform\_input, transform\_output [\#463](https://github.com/SeldonIO/seldon-core/pull/463) ([cliveseldon](https://github.com/cliveseldon))
- Script to create Seldon API testing files from any Pandas dataframe [\#460](https://github.com/SeldonIO/seldon-core/pull/460) ([Love-R](https://github.com/Love-R))
- WIP: Python wrappers rewrite [\#457](https://github.com/SeldonIO/seldon-core/pull/457) ([cliveseldon](https://github.com/cliveseldon))
- Python builder [\#455](https://github.com/SeldonIO/seldon-core/pull/455) ([gsunner](https://github.com/gsunner))
- Update redis [\#446](https://github.com/SeldonIO/seldon-core/pull/446) ([naseemkullah](https://github.com/naseemkullah))
- WIP: Autoscaling [\#437](https://github.com/SeldonIO/seldon-core/pull/437) ([cliveseldon](https://github.com/cliveseldon))

## [v0.2.6](https://github.com/SeldonIO/seldon-core/tree/v0.2.6) (2019-02-22)
[Full Changelog](https://github.com/SeldonIO/seldon-core/compare/v0.2.5...v0.2.6)

**Fixed bugs:**

- Bug parsing boolean predictive unit params in Python wrappers [\#439](https://github.com/SeldonIO/seldon-core/issues/439)
- APIFE fails to connect to service due to name change [\#433](https://github.com/SeldonIO/seldon-core/issues/433)

**Closed issues:**

- If building a python image from a folder, which is also a git-folder build silently fails [\#452](https://github.com/SeldonIO/seldon-core/issues/452)
- Setting `engineResources` not enabling resource requests/limits to `seldon-container-engine` sidecar [\#398](https://github.com/SeldonIO/seldon-core/issues/398)
- Expose Jaeger agent port as environment variable on deployment manifest [\#396](https://github.com/SeldonIO/seldon-core/issues/396)
- Ksonnets for Seldon Analytics [\#391](https://github.com/SeldonIO/seldon-core/issues/391)
- sklearn iris returns value error [\#389](https://github.com/SeldonIO/seldon-core/issues/389)
- SOAP API [\#387](https://github.com/SeldonIO/seldon-core/issues/387)
- unable to find proto file which defines grpc [\#384](https://github.com/SeldonIO/seldon-core/issues/384)
- tensorflow-gpu [\#380](https://github.com/SeldonIO/seldon-core/issues/380)
- onnx\_resnet50.ipynb : "Unknown operation: Gather" [\#379](https://github.com/SeldonIO/seldon-core/issues/379)
- Passing arguments to the model object [\#377](https://github.com/SeldonIO/seldon-core/issues/377)
- Model pod enters in CrashLoopBackOff. How to debug? [\#376](https://github.com/SeldonIO/seldon-core/issues/376)
- Global metrics show N/A in Seldon Analytics Grafana [\#371](https://github.com/SeldonIO/seldon-core/issues/371)
- Mistyped check causing NULL Pointer Exceptions with getNamespace [\#367](https://github.com/SeldonIO/seldon-core/issues/367)
- Json payload size increases when I use json.dumps [\#365](https://github.com/SeldonIO/seldon-core/issues/365)
- Need an updated tutorial for seldon serving on GKE [\#361](https://github.com/SeldonIO/seldon-core/issues/361)
- Hi,we need Golang Deploy Seldon Wrapper Container [\#356](https://github.com/SeldonIO/seldon-core/issues/356)
- Update docs and examples to use the new Python package [\#347](https://github.com/SeldonIO/seldon-core/issues/347)
- Potential problem in EpsilonGreedy.py? [\#336](https://github.com/SeldonIO/seldon-core/issues/336)
- Deploying seldon-core to Kubernetes 1.8.6 fails with `no matches for kind "Deployment" in version "apps/v1"` [\#333](https://github.com/SeldonIO/seldon-core/issues/333)
- S2i build image with private pip repository [\#330](https://github.com/SeldonIO/seldon-core/issues/330)
- Wrapping components outside of the tree [\#324](https://github.com/SeldonIO/seldon-core/issues/324)
- Seems to be a bad fit for a multi-tenant cluster. [\#308](https://github.com/SeldonIO/seldon-core/issues/308)
- Update Grafana / Prometheus image [\#303](https://github.com/SeldonIO/seldon-core/issues/303)
- Function to pass additional meta info for `predict\(\)` [\#297](https://github.com/SeldonIO/seldon-core/issues/297)
- Update base java image [\#289](https://github.com/SeldonIO/seldon-core/issues/289)
- Update ksonnet to reflect latest helm templates [\#282](https://github.com/SeldonIO/seldon-core/issues/282)
- NullPointerException in seldon-cluster manager logs [\#268](https://github.com/SeldonIO/seldon-core/issues/268)
- requestPath picking up old model on rolling update [\#267](https://github.com/SeldonIO/seldon-core/issues/267)
- Seldon deployment success/failure condition [\#255](https://github.com/SeldonIO/seldon-core/issues/255)
- Reconcile the differences between seldon-core and kubeflow core.libsonnet to improve maintenance [\#237](https://github.com/SeldonIO/seldon-core/issues/237)
- Make the "apiVersion" in the Helm templates consistent [\#236](https://github.com/SeldonIO/seldon-core/issues/236)
- Create initial docs for Transformers [\#229](https://github.com/SeldonIO/seldon-core/issues/229)
- Create initial docs for Routers [\#228](https://github.com/SeldonIO/seldon-core/issues/228)
- deploy docker image is ok ,but deploy  k8s pod  always  failed [\#212](https://github.com/SeldonIO/seldon-core/issues/212)
- Format of the data sent as a request to the seldon REST api? [\#193](https://github.com/SeldonIO/seldon-core/issues/193)
- There is no setting that allows increasing the limits of GRPC Server  [\#183](https://github.com/SeldonIO/seldon-core/issues/183)
- Docker image build error with sklearn\_iris\_docker example [\#164](https://github.com/SeldonIO/seldon-core/issues/164)
- Add support for spring-boot-starter-webflux [\#152](https://github.com/SeldonIO/seldon-core/issues/152)
- gRPC query waits indefinitely while execution giving no output [\#149](https://github.com/SeldonIO/seldon-core/issues/149)
- scikit-learn support for predict method not only predict\_proba [\#145](https://github.com/SeldonIO/seldon-core/issues/145)
- Wrapper command on windows PS [\#134](https://github.com/SeldonIO/seldon-core/issues/134)
- Error 401 while requesting prediction outputs from seldon server [\#122](https://github.com/SeldonIO/seldon-core/issues/122)
- How to Deploy our custom models on seldon-core [\#104](https://github.com/SeldonIO/seldon-core/issues/104)
- Create docs for available plugins [\#100](https://github.com/SeldonIO/seldon-core/issues/100)
- Custom model endpoints [\#96](https://github.com/SeldonIO/seldon-core/issues/96)
- Docker image missing for Iris classification [\#91](https://github.com/SeldonIO/seldon-core/issues/91)
- Add options to populate meta data in wrappers foreach API request [\#86](https://github.com/SeldonIO/seldon-core/issues/86)
- Add InputOutputTransformer predictive unit [\#85](https://github.com/SeldonIO/seldon-core/issues/85)
- Add Explainer as transformer component [\#84](https://github.com/SeldonIO/seldon-core/issues/84)
- Create wrapper for PyTorch models [\#82](https://github.com/SeldonIO/seldon-core/issues/82)
- Graph with epsilon greedy router sometimes fails on first request [\#80](https://github.com/SeldonIO/seldon-core/issues/80)
- Create integration testing script [\#73](https://github.com/SeldonIO/seldon-core/issues/73)
- Allow engine resource requests for engine to be configurable in proto definition for CRD [\#70](https://github.com/SeldonIO/seldon-core/issues/70)
- Create Concept Drift Alert Plugin [\#56](https://github.com/SeldonIO/seldon-core/issues/56)
- Create Outlier Detection Plugin [\#55](https://github.com/SeldonIO/seldon-core/issues/55)
- Create Multi-Armed Bandit Router Plugin\(s\) [\#54](https://github.com/SeldonIO/seldon-core/issues/54)
- Update docs for sklearn\_iris\_docker [\#52](https://github.com/SeldonIO/seldon-core/issues/52)
- Response should  contain indication of which predictor was used [\#50](https://github.com/SeldonIO/seldon-core/issues/50)
- Add git hooks for validation of notebooks before commit [\#10](https://github.com/SeldonIO/seldon-core/issues/10)
- Update docs and examples to illustrate complex runtime graphs [\#1](https://github.com/SeldonIO/seldon-core/issues/1)

**Merged pull requests:**

- openvino ensemble adjustments [\#444](https://github.com/SeldonIO/seldon-core/pull/444) ([dtrawins](https://github.com/dtrawins))
- Update image names for openvino demo [\#442](https://github.com/SeldonIO/seldon-core/pull/442) ([cliveseldon](https://github.com/cliveseldon))
- Fix bug in parsing boolean params in Python wrapper [\#441](https://github.com/SeldonIO/seldon-core/pull/441) ([jklaise](https://github.com/jklaise))
- Update java wrapper version in docs [\#436](https://github.com/SeldonIO/seldon-core/pull/436) ([cliveseldon](https://github.com/cliveseldon))
- Fix API Gateway Endpoint name [\#435](https://github.com/SeldonIO/seldon-core/pull/435) ([cliveseldon](https://github.com/cliveseldon))
- Updates for openvino demo [\#431](https://github.com/SeldonIO/seldon-core/pull/431) ([cliveseldon](https://github.com/cliveseldon))
- updated ensemble pipeline with OpenVINO component [\#430](https://github.com/SeldonIO/seldon-core/pull/430) ([dtrawins](https://github.com/dtrawins))
- Outlier service type [\#428](https://github.com/SeldonIO/seldon-core/pull/428) ([arnaudvl](https://github.com/arnaudvl))
- Engine merge meta puid [\#424](https://github.com/SeldonIO/seldon-core/pull/424) ([jklaise](https://github.com/jklaise))
- Allow reusing containers in the inference graph [\#423](https://github.com/SeldonIO/seldon-core/pull/423) ([jklaise](https://github.com/jklaise))
- Ambassador Update: Canary, Shadow, Header Based Routing [\#409](https://github.com/SeldonIO/seldon-core/pull/409) ([cliveseldon](https://github.com/cliveseldon))
- Cluster Manager Cache Fix [\#408](https://github.com/SeldonIO/seldon-core/pull/408) ([cliveseldon](https://github.com/cliveseldon))
- Add ability to fetch metadata from model and transformer components [\#407](https://github.com/SeldonIO/seldon-core/pull/407) ([jklaise](https://github.com/jklaise))
- Fix api-tester not working via GRPC and Ambassador [\#405](https://github.com/SeldonIO/seldon-core/pull/405) ([jklaise](https://github.com/jklaise))
- Fix s2i builder image local build to use latest Python source code [\#404](https://github.com/SeldonIO/seldon-core/pull/404) ([jklaise](https://github.com/jklaise))
- setPredictorSpec is not needed anymore in EnginePredictor [\#401](https://github.com/SeldonIO/seldon-core/pull/401) ([ro7m](https://github.com/ro7m))
- Fix engine resources setting and update docs [\#400](https://github.com/SeldonIO/seldon-core/pull/400) ([cliveseldon](https://github.com/cliveseldon))
- Provide Ksonnet Analytics Package [\#399](https://github.com/SeldonIO/seldon-core/pull/399) ([cliveseldon](https://github.com/cliveseldon))
- Allow JAEGER\_AGENT\_PORT env on default Jaeger configuration [\#397](https://github.com/SeldonIO/seldon-core/pull/397) ([masroorhasan](https://github.com/masroorhasan))
- Outlier update [\#395](https://github.com/SeldonIO/seldon-core/pull/395) ([arnaudvl](https://github.com/arnaudvl))
- removing resttemplate setter from predictionService [\#393](https://github.com/SeldonIO/seldon-core/pull/393) ([ro7m](https://github.com/ro7m))
- gRPC load balancing via Ambassador [\#390](https://github.com/SeldonIO/seldon-core/pull/390) ([cliveseldon](https://github.com/cliveseldon))
- Outlier mahalanobis [\#388](https://github.com/SeldonIO/seldon-core/pull/388) ([arnaudvl](https://github.com/arnaudvl))
- Update ngraph s2i image and remove torch from demo [\#386](https://github.com/SeldonIO/seldon-core/pull/386) ([cliveseldon](https://github.com/cliveseldon))
- ojAlgo upgrade to v47, and a few improvements [\#385](https://github.com/SeldonIO/seldon-core/pull/385) ([apete](https://github.com/apete))
- Cicd demo - WIP [\#382](https://github.com/SeldonIO/seldon-core/pull/382) ([gsunner](https://github.com/gsunner))
- Add docs for parameters in components [\#381](https://github.com/SeldonIO/seldon-core/pull/381) ([cliveseldon](https://github.com/cliveseldon))
- WIP: Update ksonnet to ensure 1.8 k8s compatibility [\#375](https://github.com/SeldonIO/seldon-core/pull/375) ([cliveseldon](https://github.com/cliveseldon))
- seq2seq lstm outlier detector [\#374](https://github.com/SeldonIO/seldon-core/pull/374) ([arnaudvl](https://github.com/arnaudvl))
- Adding test case for SeldonDeploymentWatcher [\#373](https://github.com/SeldonIO/seldon-core/pull/373) ([ro7m](https://github.com/ro7m))
- Add @Timed to 2 main REST endpoint to readd prometheus metrics [\#372](https://github.com/SeldonIO/seldon-core/pull/372) ([cliveseldon](https://github.com/cliveseldon))
- Update OpenVINO example for raw image bytes [\#370](https://github.com/SeldonIO/seldon-core/pull/370) ([cliveseldon](https://github.com/cliveseldon))
- Fix debug logging in case study files [\#369](https://github.com/SeldonIO/seldon-core/pull/369) ([jklaise](https://github.com/jklaise))
- WIP: Train on Sagemaker, Deploy on Seldon Core [\#368](https://github.com/SeldonIO/seldon-core/pull/368) ([cliveseldon](https://github.com/cliveseldon))
- Mistyped check causing NULL Pointer Exceptions with getNamespace function [\#366](https://github.com/SeldonIO/seldon-core/pull/366) ([ro7m](https://github.com/ro7m))
- Fix URICache bug in engine [\#364](https://github.com/SeldonIO/seldon-core/pull/364) ([cliveseldon](https://github.com/cliveseldon))
- release notes 0.2.5 [\#363](https://github.com/SeldonIO/seldon-core/pull/363) ([cliveseldon](https://github.com/cliveseldon))
- add README files to outlier detectors [\#362](https://github.com/SeldonIO/seldon-core/pull/362) ([arnaudvl](https://github.com/arnaudvl))
- Fix incorrect links in router docs [\#360](https://github.com/SeldonIO/seldon-core/pull/360) ([jklaise](https://github.com/jklaise))
- Mlflow Example [\#359](https://github.com/SeldonIO/seldon-core/pull/359) ([cliveseldon](https://github.com/cliveseldon))
- Initial Go Wrapper Example for Seldon Core [\#358](https://github.com/SeldonIO/seldon-core/pull/358) ([cliveseldon](https://github.com/cliveseldon))
- Distributed Tracing, Profiling docs and OpenVINO Demo \(WIP\) [\#357](https://github.com/SeldonIO/seldon-core/pull/357) ([cliveseldon](https://github.com/cliveseldon))
- Change mean\_classifier to mock\_classifier in tests and example for consistency [\#355](https://github.com/SeldonIO/seldon-core/pull/355) ([cliveseldon](https://github.com/cliveseldon))
- Update CRDs to correct OpenAPISchema [\#354](https://github.com/SeldonIO/seldon-core/pull/354) ([cliveseldon](https://github.com/cliveseldon))
- Remove legacy testers [\#352](https://github.com/SeldonIO/seldon-core/pull/352) ([jklaise](https://github.com/jklaise))
- Update example models to use python package [\#351](https://github.com/SeldonIO/seldon-core/pull/351) ([cliveseldon](https://github.com/cliveseldon))
- WIP: Update docs and examples to use Python package [\#348](https://github.com/SeldonIO/seldon-core/pull/348) ([jklaise](https://github.com/jklaise))
- S2i 0.4 update [\#346](https://github.com/SeldonIO/seldon-core/pull/346) ([jklaise](https://github.com/jklaise))
- Fix bug creating tf protos for e2e testing [\#345](https://github.com/SeldonIO/seldon-core/pull/345) ([jklaise](https://github.com/jklaise))
- Python release version [\#344](https://github.com/SeldonIO/seldon-core/pull/344) ([jklaise](https://github.com/jklaise))
- multi-armed bandit components [\#335](https://github.com/SeldonIO/seldon-core/pull/335) ([jklaise](https://github.com/jklaise))
- Enable support for using local Python binaries when wrapping components [\#332](https://github.com/SeldonIO/seldon-core/pull/332) ([jklaise](https://github.com/jklaise))
- Update build scripts to use latest core builder image [\#313](https://github.com/SeldonIO/seldon-core/pull/313) ([jklaise](https://github.com/jklaise))

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