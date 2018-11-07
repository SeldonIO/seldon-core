// @apiVersion 0.1
// @name io.ksonnet.pkg.seldon-core
// @description Seldon Core components. Operator and API FrontEnd.
// @shortDescription Seldon Core components.
// @param name string seldon Name to give seldon
// @optionalParam namespace string null Namespace to use for the components. It is automatically inherited from the environment if not set.
// @optionalParam withRbac string true Whether to include RBAC setup
// @optionalParam withApife string true Whether to include builtin API OAuth gateway server for ingress
// @optionalParam withAmbassador string false Whether to include Ambassador reverse proxy
// @optionalParam apifeServiceType string NodePort API Front End Service Type
// @optionalParam operatorSpringOpts string null cluster manager spring opts
// @optionalParam operatorJavaOpts string null cluster manager java opts
// @optionalParam grpcMaxMessageSize string 4194304 Max gRPC message size
// @optionalParam seldonVersion string 0.2.4-SNAPSHOT Seldon version
// @optionalParam engineServiceAccount string default Service account for Seldon Service Orchestrator Engine

local k = import "k.libsonnet";
local core = import "seldon-core/seldon-core/core.libsonnet";

// updatedParams uses the environment namespace if
// the namespace parameter is not explicitly set
local updatedParams = params {
  namespace: if params.namespace == "null" then env.namespace else params.namespace,
};

local seldonVersion = import "param://seldonVersion";

local name = import "param://name";
local namespace = updatedParams.namespace;
local withRbac = import "param://withRbac";
local withApife = import "param://withApife";
local withAmbassador = import "param://withAmbassador";

// APIFE
local apifeImage = "seldonio/apife:" + seldonVersion;
local apifeServiceType = import "param://apifeServiceType";
local grpcMaxMessageSize = import "param://grpcMaxMessageSize";

// Cluster Manager (The CRD Operator)
local operatorImage = "seldonio/cluster-manager:" + seldonVersion;
local operatorSpringOptsParam = import "param://operatorSpringOpts";
local operatorSpringOpts = if operatorSpringOptsParam != "null" then operatorSpringOptsParam else "";
local operatorJavaOptsParam = import "param://operatorJavaOpts";
local operatorJavaOpts = if operatorJavaOptsParam != "null" then operatorJavaOptsParam else "";

// Engine
local engineImage = "seldonio/engine:" + seldonVersion;
local engineServiceAccount = import "param://engineServiceAccount";

// APIFE
local apife = [
  core.parts(name, namespace, seldonVersion).apife(apifeImage, withRbac, grpcMaxMessageSize),
  core.parts(name, namespace, seldonVersion).apifeService(apifeServiceType),
];

local rbac2 = [
  core.parts(name, namespace, seldonVersion).rbacServiceAccount(),
  core.parts(name, namespace, seldonVersion).rbacClusterRole(),
  core.parts(name, namespace, seldonVersion).rbacRole(),
  core.parts(name, namespace, seldonVersion).rbacRoleBinding(),
  core.parts(name, namespace, seldonVersion).rbacClusterRoleBinding(),
];

local rbac1 = [
  core.parts(name, namespace, seldonVersion).rbacServiceAccount(),
  core.parts(name, namespace, seldonVersion).rbacRoleBinding(),
];

local rbac = if std.startsWith(seldonVersion, "0.1") then rbac1 else rbac2;

// Core
local coreComponents = [
  core.parts(name, namespace, seldonVersion).deploymentOperator(engineImage, operatorImage, operatorSpringOpts, operatorJavaOpts, withRbac, engineServiceAccount),
  core.parts(name, namespace, seldonVersion).redisDeployment(),
  core.parts(name, namespace, seldonVersion).redisService(),
  core.parts(name, namespace, seldonVersion).crd(),
];

//Ambassador
local ambassadorRbac = [
  core.parts(name,namespace, seldonVersion).rbacAmbassadorRole(),
  core.parts(name,namespace, seldonVersion).rbacAmbassadorRoleBinding(),  
];

local ambassador = [
  core.parts(name,namespace, seldonVersion).ambassadorDeployment(),
  core.parts(name,namespace, seldonVersion).ambassadorService(),  
];

local l1 = if withRbac == "true" then rbac + coreComponents else coreComponents;
local l2 = if withApife == "true" then l1 + apife else l1;
local l3 = if withAmbassador == "true" && withRbac == "true" then l2 + ambassadorRbac else l2;
local l4 = if withAmbassador == "true" then l3 + ambassador else l3;

l4