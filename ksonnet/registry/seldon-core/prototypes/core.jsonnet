// @apiVersion 0.1
// @name io.ksonnet.pkg.seldon-core
// @description Seldon Core components. Operator and API FrontEnd.
// @shortDescription Seldon Core components.
// @param name string Name seldon-core to give seldon-core
// @optionalParam namespace string default Namespace
// @optionalParam apifeImage string seldonio/apife:0.1.4-SNAPSHOT Default image for API Front End
// @optionalParam serviceType string LoadBalancer API Front End Service Type
// @optionalParam engineImage string seldonio/engine:0.1.4-SNAPSHOT Seldon engine image version
// @optionalParam clusterManagerImage string seldonio/cluster-manager:0.1.4-SNAPSHOT Seldon cluster manager image version
// @optionalParam springOpts string null cluster manager spring opts
// @optionalParam javaOpts string null cluster manager java opts
// @optionalParam withRbac string false Whether to include RBAC setup
// @optionalParam withApife string true Whether to include builtin API Oauth fornt end server for ingress

// TODO(https://github.com/ksonnet/ksonnet/issues/222): We have to add namespace as an explicit parameter
// because ksonnet doesn't support inheriting it from the environment yet.

local k = import 'k.libsonnet';
local core = import "seldon-core/seldon-core/core.libsonnet";

local name = import 'param://name';
local namespace = import 'param://namespace';
local apifeImage = import 'param://apifeImage';
local serviceType = import 'param://serviceType';
local withRbac = import 'param://withRbac';
local withApife = import 'param://withApife';

local engineImage = import 'param://engineImage';
local clusterManagerImage = import 'param://clusterManagerImage';
local springOptsParam = import 'param://springOpts';
local springOpts = if springOptsParam != "null" then springOptsParam else null;
local javaOptsParam = import 'param://javaOpts';
local javaOpts = if javaOptsParam != "null" then javaOptsParam else null;

//std.prune(k.core.v1.list.new(

  // seldon-core components
local apife = if withRbac == "true" then
       [core.parts(namespace).apifeWithRbac(apifeImage),core.parts(namespace).apifeService(serviceType),core.parts(namespace).rbacServiceAccount(),core.parts(namespace).rbacClusterRoleBinding()]
    else
       [core.parts(namespace).apife(apifeImage),core.parts(namespace).apifeService(serviceType)];


local coreComponents =  [
  core.parts(namespace).deploymentOperator(engineImage, clusterManagerImage, springOpts, javaOpts),
  core.parts(namespace).redisDeployment(),  
  core.parts(namespace).redisService(),
  core.parts(namespace).crd(),
];



if withApife == "true" then
k.core.v1.list.new(apife + coreComponents)
else
k.core.v1.list.new(coreComponents)

//k.core.v1.list.new([
//  apife,
//  k.core.v1.list.new([
//    core.parts(namespace).apifeService(serviceType),
//    core.parts(namespace).deploymentOperator(engineImage, clusterManagerImage, springOpts, javaOpts),
//    core.parts(namespace).redisDeployment(),  
//    core.parts(namespace).redisService(),])
//])
