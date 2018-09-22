// @apiVersion 0.1
// @name io.ksonnet.pkg.seldon-core
// @description Seldon Core components. Operator and API FrontEnd.
// @shortDescription Seldon Core components.
// @param name string seldon Name to give seldon
// @optionalParam namespace string null Namespace to use for the components. It is automatically inherited from the environment if not set.
// @optionalParam withRbac string true Whether to include RBAC setup
// @optionalParam withApife string true Whether to include builtin API OAuth gateway server for ingress
// @optionalParam withAmbassador string false Whether to include Ambassador reverse proxy
// @optionalParam apifeImage string seldonio/apife:0.2.4-SNAPSHOT Default image for API Front End
// @optionalParam apifeServiceType string NodePort API Front End Service Type
// @optionalParam operatorImage string seldonio/cluster-manager:0.2.4-SNAPSHOT Seldon cluster manager image version
// @optionalParam operatorSpringOpts string null cluster manager spring opts
// @optionalParam operatorJavaOpts string null cluster manager java opts
// @optionalParam engineImage string seldonio/engine:0.2.4-SNAPSHOT Seldon engine image version
// @optionalParam grpcMaxMessageSize string 4194304 Max gRPC message size

local k = import "k.libsonnet";
local core = import "seldon-core/seldon-core/core.libsonnet";

// updatedParams uses the environment namespace if
// the namespace parameter is not explicitly set
local updatedParams = params {
  namespace: if params.namespace == "null" then env.namespace else params.namespace,
};

local name = import "param://name";
local namespace = updatedParams.namespace;
local withRbac = import "param://withRbac";
local withApife = import "param://withApife";
local withAmbassador = import "param://withAmbassador";

// APIFE
local apifeImage = import "param://apifeImage";
local apifeServiceType = import "param://apifeServiceType";
local grpcMaxMessageSize = import "param://grpcMaxMessageSize";

// Cluster Manager (The CRD Operator)
local operatorImage = import "param://operatorImage";
local operatorSpringOptsParam = import "param://operatorSpringOpts";
local operatorSpringOpts = if operatorSpringOptsParam != "null" then operatorSpringOptsParam else "";
local operatorJavaOptsParam = import "param://operatorJavaOpts";
local operatorJavaOpts = if operatorJavaOptsParam != "null" then operatorJavaOptsParam else "";

// Engine
local engineImage = import "param://engineImage";

// APIFE
local apife = [
  core.parts(name,namespace).apife(apifeImage, withRbac, grpcMaxMessageSize),
  core.parts(name,namespace).apifeService(apifeServiceType),
];

local rbac = [
  core.parts(name,namespace).rbacServiceAccount(),
  core.parts(name,namespace).rbacClusterRole(),
  core.parts(name,namespace).rbacRole(),
  core.parts(name,namespace).rbacRoleBinding(),  
  core.parts(name,namespace).rbacClusterRoleBinding(),
];


// Core
local coreComponents = [
  core.parts(name,namespace).deploymentOperator(engineImage, operatorImage, operatorSpringOpts, operatorJavaOpts, withRbac),
  core.parts(name,namespace).redisDeployment(),
  core.parts(name,namespace).redisService(),
  core.parts(name,namespace).crd(),
];


//Ambassador
local ambassadorRbac = [
  core.parts(name,namespace).rbacAmbassadorRole(),
  core.parts(name,namespace).rbacAmbassadorRoleBinding(),  
];

local ambassador = [
  core.parts(name,namespace).ambassadorDeployment(),
  core.parts(name,namespace).ambassadorService(),  
];

local l1 = if withRbac == "true" then rbac + coreComponents else coreComponents;
local l2 = if withApife == "true" then l1 + apife else l1;
local l3 = if withAmbassador == "true" && withRbac == "true" then l2 + ambassadorRbac else l2;
local l4 = if withAmbassador == "true" then l3 + ambassador else l3;

l4