// @apiVersion 0.1
// @name io.ksonnet.pkg.seldon-core-analytics
// @description Seldon Core example Grafana and Prometheus installation
// @shortDescription Seldon Core Grafana amd Prometheus with example dashboards.
// @param name string seldon Name to give seldon analytics
// @optionalParam namespace string null Namespace to use for the components. It is automatically inherited from the environment if not set.
// @optionalParam password string admin Admin password for Grafana
// @optionalParam prometheusServiceType string ClusterIP Prometheus service type
// @optionalParam grafanaServiceType string NodePort Grafana service type

local k = import "k.libsonnet";
local analytics = import "seldon-core/seldon-core-analytics/analytics.libsonnet";

// updatedParams uses the environment namespace if
// the namespace parameter is not explicitly set
local updatedParams = params {
  namespace: if params.namespace == "null" then env.namespace else params.namespace,
};

local name = import "param://name";
local namespace = updatedParams.namespace;
local password = import "param://password";
local prometheusServiceType = import "param://prometheusServiceType";
local grafanaServiceType = import "param://grafanaServiceType";

// Analytics Resources
local resources = [
  analytics.parts(name, namespace).grafanaPromSecret(password),
  analytics.parts(name, namespace).alertManagerServerConf(),
  analytics.parts(name, namespace).grafanaImportDashboards(),
  analytics.parts(name, namespace).prometheusRules(),
  analytics.parts(name, namespace).prometheusServerConf(),
  analytics.parts(name, namespace).prometheusClusterRole(),
  analytics.parts(name, namespace).prometheusServiceAccount(),
  analytics.parts(name, namespace).prometheusClusterRoleBinding(),
  analytics.parts(name, namespace).alertManagerDeployment(),
  analytics.parts(name, namespace).alertManagerService(),
  analytics.parts(name, namespace).grafanaPromDeployment(),
  analytics.parts(name, namespace).grafanaPromService(grafanaServiceType),
  analytics.parts(name, namespace).grafanaPromJob(),
  analytics.parts(name, namespace).prometheusExporter(),
  analytics.parts(name, namespace).prometheusExporterService(),
  analytics.parts(name, namespace).prometheusDeployment(),
  analytics.parts(name, namespace).prometheusService(prometheusServiceType),    
];

resources
