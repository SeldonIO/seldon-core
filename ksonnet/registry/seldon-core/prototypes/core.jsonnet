// @apiVersion 0.1
// @name io.ksonnet.pkg.seldon-core
// @description Seldon Core components. Operator and API FrontEnd.
// @shortDescription Seldon Core components.
// @param name string Name seldon-core to give seldon-core
// @optionalParam namespace string default Namespace
// @optionalParam withRbac string false Whether to include RBAC setup
// @optionalParam withApife string true Whether to include builtin API Oauth fornt end server for ingress
// @optionalParam apifeImage string seldonio/apife:0.1.4-SNAPSHOT Default image for API Front End
// @optionalParam apifeServiceType string LoadBalancer API Front End Service Type
// @optionalParam operatorImage string seldonio/cluster-manager:0.1.4-SNAPSHOT Seldon cluster manager image version
// @optionalParam operatorSpringOpts string null cluster manager spring opts
// @optionalParam operatorJavaOpts string null cluster manager java opts
// @optionalParam engineImage string seldonio/engine:0.1.4-SNAPSHOT Seldon engine image version

// TODO(https://github.com/ksonnet/ksonnet/issues/222): We have to add namespace as an explicit parameter
// because ksonnet doesn't support inheriting it from the environment yet.

local k = import 'k.libsonnet';
local core = import "seldon-core/seldon-core/core.libsonnet";

local name = import 'param://name';
local namespace = import 'param://namespace';
local withRbac = import 'param://withRbac';
local withApife = import 'param://withApife';

// APIFE
local apifeImage = import 'param://apifeImage';
local apifeServiceType = import 'param://apifeServiceType';

// Cluster Manager (The CRD Operator)
local operatorImage = import 'param://operatorImage';
local operatorSpringOptsParam = import 'param://operatorSpringOpts';
local operatorSpringOpts = if operatorSpringOptsParam != "null" then operatorSpringOptsParam else "";
local operatorJavaOptsParam = import 'param://operatorJavaOpts';
local operatorJavaOpts = if operatorJavaOptsParam != "null" then operatorJavaOptsParam else "";

// Engine
local engineImage = import 'param://engineImage';

// APIFE
local apife = [
  core.parts(namespace).apife(apifeImage, withRbac),
  core.parts(namespace).apifeService(apifeServiceType),
];

local rbac = [
  core.parts(namespace).rbacServiceAccount(),
  core.parts(namespace).rbacClusterRoleBinding(),
];

// Core
local coreComponents = [
  core.parts(namespace).deploymentOperator(engineImage, operatorImage, operatorSpringOpts, operatorJavaOpts, withRbac),
  core.parts(namespace).redisDeployment(),
  core.parts(namespace).redisService(),
  core.parts(namespace).crd(),
];

if withRbac == "true" && withApife == "true" then
k.core.v1.list.new(apife + rbac + coreComponents)
else if withRbac == "true" && withApife == "false" then
k.core.v1.list.new(rbac + coreComponents)
else if withRbac == "false" && withApife == "true" then
k.core.v1.list.new(apife + coreComponents)
else if withRbac == "false" && withApife == "false" then
k.core.v1.list.new(coreComponents)
