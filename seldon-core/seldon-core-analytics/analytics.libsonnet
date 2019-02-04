local k = import "k.libsonnet";
local deployment = k.extensions.v1beta1.deployment;
local container = k.apps.v1beta1.deployment.mixin.spec.template.spec.containersType;
local service = k.core.v1.service.mixin;
local serviceAccountMixin = k.core.v1.serviceAccount.mixin;
local clusterRoleBindingMixin = k.rbac.v1beta1.clusterRoleBinding.mixin;
local clusterRoleBinding = k.rbac.v1beta1.clusterRoleBinding;
local roleBindingMixin = k.rbac.v1beta1.roleBinding.mixin;
local roleBinding = k.rbac.v1beta1.roleBinding;
local roleMixin = k.rbac.v1beta1.role.mixin;
local serviceAccount = k.core.v1.serviceAccount;
local secret = k.core.v1.secret;
local configMap = k.core.v1.configMap;
local job = k.batch.v1.job;
local daemonSet = k.extensions.v1beta1.daemonSet;

local analyticsTemplate = import "json/analytics.json";

local getGrafanaPromSecret(x) = std.endsWith(x.metadata.name, "grafana-prom-secret");
local getAlertmanagerServerConf(x) = std.endsWith(x.metadata.name, "alertmanager-server-conf") && x.kind == "ConfigMap";
local getGrafanaImportDashboards(x) = std.endsWith(x.metadata.name, "grafana-import-dashboards") && x.kind == "ConfigMap";
local getPrometheusRules(x) = std.endsWith(x.metadata.name, "prometheus-rules") && x.kind == "ConfigMap";
local getPrometheusServerConf(x) = std.endsWith(x.metadata.name, "prometheus-server-conf") && x.kind == "ConfigMap";
local getPrometheusClusterRole(x) = x.metadata.name == "prometheus" && x.kind == "ClusterRole";
local getPrometheusServiceAccount(x) = x.metadata.name == "prometheus" && x.kind == "ServiceAccount";
local getPrometheusClusterRoleBinding(x) = x.metadata.name == "prometheus" && x.kind == "ClusterRoleBinding";
local getAlertManagerDeployment(x) = x.metadata.name == "alertmanager-deployment" && x.kind == "Deployment";
local getAlertManagerService(x) = x.metadata.name == "alertmanager" && x.kind == "Service";
local getGrafanaPromDeployment(x) = x.metadata.name == "grafana-prom-deployment" && x.kind == "Deployment";
local getGrafanaPromService(x) = x.metadata.name == "grafana-prom" && x.kind == "Service";
local getGrafanaPromJob(x) = x.metadata.name == "grafana-prom-import-dashboards" && x.kind == "Job";
local getPrometheusExporter(x) = x.metadata.name == "prometheus-node-exporter" && x.kind == "DaemonSet";
local getPrometheusExporterService(x) = x.metadata.name == "prometheus-node-exporter" && x.kind == "Service";
local getPrometheusDeployment(x) = x.metadata.name == "prometheus-deployment" && x.kind == "Deployment";
local getPrometheusService(x) = x.metadata.name == "prometheus-seldon" && x.kind == "Service";

{
  parts(name, namespace)::

    {
      grafanaPromSecret(password)::

        local baseGrafanaPromSecret = std.filter(getGrafanaPromSecret, analyticsTemplate.items)[0];
	baseGrafanaPromSecret +
		secret.withDataMixin({"grafana-prom-admin-password": std.base64(password)}) +
		secret.mixin.metadata.withNamespace(namespace),

      alertManagerServerConf()::

        local baseAlertManagerServerConf = std.filter(getAlertmanagerServerConf, analyticsTemplate.items)[0];
	baseAlertManagerServerConf +
		configMap.mixin.metadata.withNamespace(namespace),

      grafanaImportDashboards()::

        local baseGrafanaImportDashboards = std.filter(getGrafanaImportDashboards, analyticsTemplate.items)[0];
	baseGrafanaImportDashboards +
		configMap.mixin.metadata.withNamespace(namespace),

      prometheusRules()::

        local basePrometheusRules = std.filter(getPrometheusRules, analyticsTemplate.items)[0];
	basePrometheusRules +
		configMap.mixin.metadata.withNamespace(namespace),	

      prometheusServerConf()::

        local basePrometheusServerConf = std.filter(getPrometheusServerConf, analyticsTemplate.items)[0];
	basePrometheusServerConf +
		configMap.mixin.metadata.withNamespace(namespace),		
	
      prometheusClusterRole()::

        std.filter(getPrometheusClusterRole, analyticsTemplate.items)[0],
	
      prometheusServiceAccount()::

        local basePrometheusServiceAccount = std.filter(getPrometheusServiceAccount, analyticsTemplate.items)[0];
	basePrometheusServiceAccount +
		serviceAccountMixin.metadata.withNamespace(namespace),
	
      prometheusClusterRoleBinding()::

        local rbacClusterRoleBinding = std.filter(getPrometheusClusterRoleBinding, analyticsTemplate.items)[0];

        local subject = rbacClusterRoleBinding.subjects[0]
                        { namespace: namespace };

        rbacClusterRoleBinding +
        clusterRoleBindingMixin.metadata.withNamespace(namespace) +
        clusterRoleBinding.withSubjects([subject]),


      alertManagerDeployment()::

        local baseAlertManagerDeployment = std.filter(getAlertManagerDeployment, analyticsTemplate.items)[0];
	baseAlertManagerDeployment +
		deployment.mixin.metadata.withNamespace(namespace),
	
      alertManagerService()::

        local baseAlertManagerService = std.filter(getAlertManagerService, analyticsTemplate.items)[0];
	baseAlertManagerService +
		service.metadata.withNamespace(namespace),
	
      grafanaPromDeployment()::

        local baseGrafanaDeployment = std.filter(getGrafanaPromDeployment, analyticsTemplate.items)[0];
	baseGrafanaDeployment +
		deployment.mixin.metadata.withNamespace(namespace),
		
      grafanaPromService(serviceType)::

        local baseServiceGrafana = std.filter(getGrafanaPromService, analyticsTemplate.items)[0];

	baseServiceGrafana +
		service.spec.withType(serviceType) +
		service.metadata.withNamespace(namespace),
		
      grafanaPromJob()::

        local baseGrafanaPromJob = std.filter(getGrafanaPromJob, analyticsTemplate.items)[0];
	baseGrafanaPromJob +
		job.mixin.metadata.withNamespace(namespace),
	
      prometheusExporter()::

        local basePrometheuExporter = std.filter(getPrometheusExporter, analyticsTemplate.items)[0];
	basePrometheuExporter +
		daemonSet.mixin.metadata.withNamespace(namespace),
	
      prometheusExporterService()::

        local basePrometheusExporterService = std.filter(getPrometheusExporterService, analyticsTemplate.items)[0];
	basePrometheusExporterService +
		service.metadata.withNamespace(namespace),
	
      prometheusDeployment()::

        local basePrometheusDeployment = std.filter(getPrometheusDeployment, analyticsTemplate.items)[0];
	basePrometheusDeployment +
		deployment.mixin.metadata.withNamespace(namespace),	
	
      prometheusService(serviceType)::

        local baseServiceProm = std.filter(getPrometheusService, analyticsTemplate.items)[0];

	baseServiceProm +
	 	service.spec.withType(serviceType) +
		service.metadata.withNamespace(namespace),
		
    },  // parts
}
