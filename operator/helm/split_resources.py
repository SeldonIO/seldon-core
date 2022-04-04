import argparse
import glob
import re

import yaml

parser = argparse.ArgumentParser()
parser.add_argument("--prefix", default="xx",
                    help="find files matching prefix")
parser.add_argument("--folder", required=True, help="Output folder")
args, _ = parser.parse_known_args()

HELM_SPARTAKUS_IF_START = "{{- if .Values.usageMetrics.enabled }}\n"
HELM_CRD_IF_START = "{{- if .Values.crd.create }}\n"
HELM_NOT_SINGLE_NAMESPACE_IF_START = "{{- if not .Values.singleNamespace }}\n"
HELM_SINGLE_NAMESPACE_IF_START = "{{- if .Values.singleNamespace }}\n"
HELM_CONTROLLERID_IF_START = "{{- if .Values.controllerId }}\n"
HELM_NOT_CONTROLLERID_IF_START = "{{- if not .Values.controllerId }}\n"
HELM_RBAC_IF_START = "{{- if .Values.rbac.create }}\n"
HELM_RBAC_CSS_IF_START = "{{- if .Values.rbac.configmap.create }}\n"
HELM_SA_IF_START = "{{- if .Values.serviceAccount.create -}}\n"
HELM_CERTMANAGER_IF_START = "{{- if .Values.certManager.enabled -}}\n"
HELM_NOT_CERTMANAGER_IF_START = "{{- if not .Values.certManager.enabled -}}\n"
HELM_VERSION_IF_START = (
    '{{- if semverCompare ">=1.15.0" .Capabilities.KubeVersion.GitVersion }}\n'
)
HELM_KUBEFLOW_IF_START = "{{- if .Values.kubeflow }}\n"
HELM_KUBEFLOW_IF_NOT_START = "{{- if not .Values.kubeflow }}\n"
HELM_CREATERESOURCES_IF_START = "{{- if not .Values.managerCreateResources }}\n"
HELM_CREATERESOURCES_RBAC_IF_START = "{{- if .Values.managerCreateResources }}\n"
HELM_K8S_V1_CRD_IF_START = '{{- if or (ge (int (regexFind "[0-9]+" .Capabilities.KubeVersion.Minor)) 18) (.Values.crd.forcev1) }}\n'
HELM_K8S_V1BETA1_CRD_IF_START = '{{- if or (lt (int (regexFind "[0-9]+" .Capabilities.KubeVersion.Minor)) 18) (.Values.crd.forcev1beta1) }}\n'
HELM_CRD_ANNOTATIONS_WITH_START = '{{- with .Values.crd.annotations }}\n'
HELM_ANNOTATIONS_TOYAML4 = '{{- toYaml . | nindent 4}}\n'
HELM_ANNOTATIONS_TOYAML8 = '{{- toYaml . | nindent 8}}\n'
HELM_CONTROLER_DEP_ANNOTATIONS_WITH_START = '{{- with .Values.manager.annotations }}\n'
HELM_CONTROLER_DEP_POD_SEC_CTX_WITH_START = '{{- with .Values.manager.containerSecurityContext }}\n'
HELM_IF_END = "{{- end }}\n"

HELM_ENV_SUBST = {
    "AMBASSADOR_ENABLED": "ambassador.enabled",
    "AMBASSADOR_SINGLE_NAMESPACE": "ambassador.singleNamespace",
    "ISTIO_ENABLED": "istio.enabled",
    "KEDA_ENABLED": "keda.enabled",
    "ISTIO_GATEWAY": "istio.gateway",
    "ISTIO_TLS_MODE": "istio.tlsMode",
    "PREDICTIVE_UNIT_HTTP_SERVICE_PORT": "predictiveUnit.httpPort",
    "PREDICTIVE_UNIT_GRPC_SERVICE_PORT": "predictiveUnit.grpcPort",
    "PREDICTIVE_UNIT_DEFAULT_ENV_SECRET_REF_NAME": "predictiveUnit.defaultEnvSecretRefName",
    "PREDICTIVE_UNIT_METRICS_PORT_NAME": "predictiveUnit.metricsPortName",
    "EXECUTOR_CONTAINER_IMAGE_PULL_POLICY": "executor.image.pullPolicy",
    "EXECUTOR_SERVER_PORT": "executor.port",
    "EXECUTOR_SERVER_METRICS_PORT_NAME": "executor.metricsPortName",
    "EXECUTOR_PROMETHEUS_PATH": "executor.prometheus.path",
    "EXECUTOR_CONTAINER_USER": "executor.user",
    "EXECUTOR_CONTAINER_SERVICE_ACCOUNT_NAME": "executor.serviceAccount.name",
    "MANAGER_CREATE_RESOURCES": "managerCreateResources",
    "EXECUTOR_REQUEST_LOGGER_DEFAULT_ENDPOINT": "executor.requestLogger.defaultEndpoint",
    "DEFAULT_USER_ID": "defaultUserID",
    "EXECUTOR_DEFAULT_CPU_LIMIT": "executor.resources.cpuLimit",
    "EXECUTOR_DEFAULT_CPU_REQUEST": "executor.resources.cpuRequest",
    "EXECUTOR_DEFAULT_MEMORY_LIMIT": "executor.resources.memoryLimit",
    "EXECUTOR_DEFAULT_MEMORY_REQUEST": "executor.resources.memoryRequest",
    "MANAGER_LOG_LEVEL": "manager.logLevel",
    "MANAGER_LEADER_ELECTION_ID": "manager.leaderElectionID",
    "EXECUTOR_REQUEST_LOGGER_WORK_QUEUE_SIZE": "executor.requestLogger.workQueueSize",
    "EXECUTOR_REQUEST_LOGGER_WRITE_TIMEOUT_MS": "executor.requestLogger.writeTimeoutMs",
    "DEPLOYMENT_NAME_AS_PREFIX": "manager.deploymentNameAsPrefix",
}
HELM_VALUES_IMAGE_PULL_POLICY = "{{ .Values.image.pullPolicy }}"


def helm_value(value: str):
    return "{{ .Values." + value + " }}"


def helm_value_json(value: str):
    return "{{ .Values." + value + " | toJson }}"


def helm_release(value: str):
    return "{{ .Release." + value + " }}"


def helm_namespace_override():
    return '{{ include "seldon.namespace" . }}'


if __name__ == "__main__":
    exp = args.prefix + "*"
    files = glob.glob(exp)
    webhookData = HELM_CREATERESOURCES_IF_START
    webhookData = (
        webhookData
        + '{{- $altNames := list ( printf "seldon-webhook-service.%s" (include "seldon.namespace" .) ) ( printf "seldon-webhook-service.%s.svc" (include "seldon.namespace" .) ) -}}\n'
    )
    webhookData = webhookData + \
        '{{- $ca := genCA "custom-metrics-ca" 365 -}}\n'
    webhookData = (
        webhookData
        + '{{- $cert := genSignedCert "seldon-webhook-service" nil $altNames 365 $ca -}}\n'
    )

    for file in files:
        with open(file, "r") as stream:
            res = yaml.safe_load(stream)
            kind = res["kind"].lower()
            name = res["metadata"]["name"].lower()
            version = res["apiVersion"]
            filename = args.folder + "/" + \
                (kind + "_" + name).lower() + ".yaml"
            print(filename)
            print(version)
            if (
                filename
                == (
                    args.folder
                    + "/"
                    + "customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml"
                )
                and version == "apiextensions.k8s.io/v1"
            ):
                print("MATCH")
                filename = (
                    args.folder
                    + "/"
                    + "customresourcedefinition_v1_seldondeployments.machinelearning.seldon.io.yaml"
                )

            print("Processing ", file)
            # Update common labels
            if "metadata" in res and "labels" in res["metadata"]:
                res["metadata"]["labels"][
                    "app.kubernetes.io/instance"
                ] = "{{ .Release.Name }}"
                res["metadata"]["labels"][
                    "app.kubernetes.io/name"
                ] = '{{ include "seldon.name" . }}'
                res["metadata"]["labels"][
                    "app.kubernetes.io/version"
                ] = "{{ .Chart.Version }}"

            # Update namespace to be helm var only if we are deploying into seldon-system
            if "metadata" in res and "namespace" in res["metadata"]:
                if (
                    res["metadata"]["namespace"] == "seldon-system"
                    or res["metadata"]["namespace"] == "seldon1-system"
                ):
                    res["metadata"]["namespace"] = helm_namespace_override()

            # controller manager
            if kind == "deployment" and name == "seldon-controller-manager":
                res["spec"]["template"]["spec"]["containers"][0][
                    "imagePullPolicy"
                ] = helm_value("image.pullPolicy")
                res["spec"]["template"]["spec"]["containers"][0][
                    "image"
                ] = "{{ .Values.image.registry }}/{{ .Values.image.repository }}:{{ .Values.image.tag }}"

                # ServiceAccount
                res["spec"]["template"]["spec"]["serviceAccountName"] = helm_value(
                    "serviceAccount.name"
                )
                # Security Context
                res["spec"]["template"]["spec"]["securityContext"][
                    "runAsUser"
                ] = helm_value("managerUserID")

                # Priority class name
                res["spec"]["template"]["spec"]["priorityClassName"] = helm_value("manager.priorityClassName")

                # Resource requests
                res["spec"]["template"]["spec"]["containers"][0]["resources"][
                    "requests"
                ]["cpu"] = helm_value("manager.cpuRequest")
                res["spec"]["template"]["spec"]["containers"][0]["resources"][
                    "requests"
                ]["memory"] = helm_value("manager.memoryRequest")
                res["spec"]["template"]["spec"]["containers"][0]["resources"]["limits"][
                    "cpu"
                ] = helm_value("manager.cpuLimit")
                res["spec"]["template"]["spec"]["containers"][0]["resources"]["limits"][
                    "memory"
                ] = helm_value("manager.memoryLimit")

                for env in res["spec"]["template"]["spec"]["containers"][0]["env"]:
                    if env["name"] in HELM_ENV_SUBST:
                        env["value"] = helm_value(HELM_ENV_SUBST[env["name"]])
                    elif env["name"] == "EXECUTOR_CONTAINER_IMAGE_AND_VERSION":
                        env[
                            "value"
                        ] = "{{ .Values.executor.image.registry }}/{{ .Values.executor.image.repository }}:{{ .Values.executor.image.tag }}"
                    elif env["name"] == "CONTROLLER_ID":
                        env["value"] = "{{ .Values.controllerId }}"

                # Update webhook port
                for portSpec in res["spec"]["template"]["spec"]["containers"][0][
                    "ports"
                ]:
                    if portSpec["name"] == "webhook-server":
                        portSpec["containerPort"] = helm_value("webhook.port")
                for argIdx in range(
                    0, len(res["spec"]["template"]["spec"]
                           ["containers"][0]["args"])
                ):
                    if (
                        res["spec"]["template"]["spec"]["containers"][0]["args"][argIdx]
                        == "--webhook-port=4443"
                    ):
                        res["spec"]["template"]["spec"]["containers"][0]["args"][
                            argIdx
                        ] = "--webhook-port=" + helm_value("webhook.port")
                res["spec"]["template"]["spec"]["containers"][0]["args"].append(
                    '{{- if .Values.singleNamespace }}--namespace={{ include "seldon.namespace" . }}{{- end }}'
                )

                # Update metrics port
                res["spec"]["template"]["metadata"]["annotations"]["prometheus.io/port"] = helm_value('metrics.port')
                for portSpec in res["spec"]["template"]["spec"]["containers"][0][
                    "ports"
                ]:
                    if portSpec["name"] == "metrics":
                        portSpec["containerPort"] = helm_value("metrics.port")

                # Networking
                res["spec"]["template"]["spec"]["hostNetwork"] = helm_value("hostNetwork")

            if kind == "configmap" and name == "seldon-config":
                res["data"]["credentials"] = helm_value_json("credentials")
                res["data"]["predictor_servers"] = helm_value_json(
                    "predictor_servers")
                res["data"]["storageInitializer"] = helm_value_json(
                    "storageInitializer"
                )
                res["data"]["explainer"] = helm_value_json("explainer")

            if kind == "serviceaccount" and name == "seldon-manager":
                res["metadata"]["name"] = helm_value("serviceAccount.name")

            if kind == "clusterrole":
                res["metadata"]["name"] = (
                    res["metadata"]["name"] + "-" + helm_namespace_override()
                )

            # Update cluster role bindings
            if kind == "clusterrolebinding":
                res["metadata"]["name"] = (
                    res["metadata"]["name"] + "-" + helm_namespace_override()
                )
                res["roleRef"]["name"] = (
                    res["roleRef"]["name"] + "-" + helm_namespace_override()
                )
                if name == "seldon-manager-rolebinding":
                    res["subjects"][0]["name"] = helm_value(
                        "serviceAccount.name")
                    res["subjects"][0]["namespace"] = helm_namespace_override()
                elif name != "seldon-spartakus-volunteer":
                    res["subjects"][0]["namespace"] = helm_namespace_override()

            # Update role bindings
            if kind == "rolebinding":
                res["subjects"][0]["namespace"] = helm_namespace_override()
                if (
                    name == "seldon1-manager-rolebinding"
                    or name == "seldon1-manager-sas-rolebinding"
                    or name == "seldon-leader-election-rolebinding"
                ):
                    res["subjects"][0]["name"] = helm_value(
                        "serviceAccount.name")
                    res["subjects"][0]["namespace"] = helm_namespace_override()

            # Update webhook certificates
            if name == "seldon-webhook-server-cert" and kind == "secret":
                res["data"]["ca.crt"] = "{{ $ca.Cert | b64enc }}"
                res["data"]["tls.crt"] = "{{ $cert.Cert | b64enc }}"
                res["data"]["tls.key"] = "{{ $cert.Key | b64enc }}"

            if kind == "validatingwebhookconfiguration":
                res["metadata"]["name"] = (
                    res["metadata"]["name"] + "-" + helm_namespace_override()
                )
                res["webhooks"][0]["clientConfig"][
                    "caBundle"
                ] = "{{ $ca.Cert | b64enc }}"
                res["webhooks"][0]["clientConfig"]["service"][
                    "namespace"
                ] = helm_namespace_override()
                if "cert-manager.io/inject-ca-from" in res["metadata"]["annotations"]:
                    res["metadata"]["annotations"]["cert-manager.io/inject-ca-from"] = (
                        helm_namespace_override() + "/seldon-serving-cert"
                    )

            if kind == "certificate":
                res["spec"][
                    "commonName"
                ] = '{{- printf "seldon-webhook-service.%s.svc" (include "seldon.namespace" .) -}}'
                res["spec"]["dnsNames"][
                    0
                ] = '{{- printf "seldon-webhook-service.%s.svc.cluster.local" (include "seldon.namespace" .) -}}'
                res["spec"]["dnsNames"][
                    1
                ] = '{{- printf "seldon-webhook-service.%s.svc" (include "seldon.namespace" .) -}}'

            if (
                kind == "customresourcedefinition"
                and name == "seldondeployments.machinelearning.seldon.io"
            ):
                # Will only work for cert-manager at present as caBundle would need to be generated in same file as secrets above
                if "conversion" in res["spec"]:
                    res["spec"]["conversion"]["webhookClientConfig"]["caBundle"] = "=="
                if "cert-manager.io/inject-ca-from" in res["metadata"]["annotations"]:
                    res["metadata"]["annotations"]["cert-manager.io/inject-ca-from"] = (
                        helm_namespace_override() + "/seldon-serving-cert"
                    )

            # Update webhook service port
            if kind == "service" and name == "seldon-webhook-service":
                res["spec"]["ports"][0]["targetPort"] = helm_value(
                    "webhook.port")

            fdata = yaml.dump(res, width=1000)

            # Spartatkus
            if name.find("spartakus") > -1:
                fdata = HELM_SPARTAKUS_IF_START + fdata + HELM_IF_END
            elif name == "seldon-webhook-rolebinding" or name == "seldon-webhook-role":
                fdata = (
                    HELM_CREATERESOURCES_RBAC_IF_START
                    + HELM_RBAC_IF_START
                    + fdata
                    + HELM_IF_END
                    + HELM_IF_END
                )
            # cluster roles for single namespace
            elif name == "seldon-manager-rolebinding" or name == "seldon-manager-role":
                fdata = (
                    HELM_NOT_SINGLE_NAMESPACE_IF_START
                    + HELM_RBAC_IF_START
                    + fdata
                    + HELM_IF_END
                    + HELM_IF_END
                )
            elif (
                name == "seldon-manager-sas-rolebinding"
                or name == "seldon-manager-sas-role"
            ):
                fdata = (
                    HELM_NOT_SINGLE_NAMESPACE_IF_START
                    + HELM_RBAC_IF_START
                    + HELM_RBAC_CSS_IF_START
                    + fdata
                    + HELM_IF_END
                    + HELM_IF_END
                    + HELM_IF_END
                )
            # roles/rolebindings for single namespace
            elif (
                name == "seldon1-manager-rolebinding" or name == "seldon1-manager-role"
            ):
                fdata = (
                    HELM_SINGLE_NAMESPACE_IF_START
                    + HELM_RBAC_IF_START
                    + fdata
                    + HELM_IF_END
                    + HELM_IF_END
                )
            elif (
                name == "seldon1-manager-sas-role"
                or name == "seldon1-manager-sas-rolebinding"
            ):
                fdata = (
                    HELM_SINGLE_NAMESPACE_IF_START
                    + HELM_RBAC_IF_START
                    + HELM_RBAC_CSS_IF_START
                    + fdata
                    + HELM_IF_END
                    + HELM_IF_END
                    + HELM_IF_END
                )
            # manager role binding
            elif (
                name == "seldon-manager-cm-rolebinding"
                or name == "seldon-manager-cm-role"
            ):
                fdata = (
                    HELM_RBAC_IF_START
                    + HELM_RBAC_CSS_IF_START
                    + fdata
                    + HELM_IF_END
                    + HELM_IF_END
                )
            elif (
                name == "seldon-leader-election-rolebinding"
                or name == "seldon-leader-election-role"
            ):
                fdata = HELM_RBAC_IF_START + fdata + HELM_IF_END
            elif name == "seldon-manager" and kind == "serviceaccount":
                fdata = HELM_SA_IF_START + fdata + HELM_IF_END
            elif kind == "issuer" or kind == "certificate":
                fdata = HELM_CERTMANAGER_IF_START + fdata + HELM_IF_END
            elif name == "seldon-webhook-server-cert" and kind == "secret":
                fdata = HELM_NOT_CERTMANAGER_IF_START + fdata + HELM_IF_END
            elif (
                name == "seldondeployments.machinelearning.seldon.io"
                and version == "apiextensions.k8s.io/v1beta1"
            ):
                fdata = (
                    HELM_CRD_IF_START
                    + HELM_K8S_V1BETA1_CRD_IF_START
                    + re.sub(
                        r"(.*controller-gen.kubebuilder.io/version.*\n)",
                        r"\1" + HELM_CRD_ANNOTATIONS_WITH_START +
                        HELM_ANNOTATIONS_TOYAML4 + HELM_IF_END,
                        fdata,
                        re.M,
                    )
                    + HELM_IF_END
                    + HELM_IF_END
                )
            elif (
                name == "seldondeployments.machinelearning.seldon.io"
                and version == "apiextensions.k8s.io/v1"
            ):
                fdata = (
                    HELM_CRD_IF_START
                    + HELM_K8S_V1_CRD_IF_START
                    + re.sub(
                        r"(.*controller-gen.kubebuilder.io/version.*\n)",
                        r"\1" + HELM_CRD_ANNOTATIONS_WITH_START +
                        HELM_ANNOTATIONS_TOYAML4 + HELM_IF_END,
                        fdata,
                        re.M,
                    )
                    + HELM_IF_END
                    + HELM_IF_END
                )
            elif kind == "service" and name == "seldon-webhook-service":
                fdata = HELM_CREATERESOURCES_IF_START + fdata + HELM_IF_END
            elif kind == "configmap" and name == "seldon-config":
                fdata = HELM_CREATERESOURCES_IF_START + fdata + HELM_IF_END
            elif kind == "deployment" and name == "seldon-controller-manager":
                fdata = re.sub(
                    r"(.*template:\n.*metadata:\n.*annotations:\n)",
                    r"\1" + HELM_CONTROLER_DEP_ANNOTATIONS_WITH_START +
                    HELM_ANNOTATIONS_TOYAML8 + HELM_IF_END,
                    fdata,
                    re.M,
                )

                fdata = re.sub(
                    r"(.*volumeMounts:\n.*\n.*\n.*\n)",
                    HELM_CREATERESOURCES_IF_START + r"\1" + HELM_IF_END,
                    fdata,
                    re.M,
                )
                fdata = re.sub(
                    r"(.*volumes:\n.*\n.*\n.*\n.*\n)",
                    HELM_CREATERESOURCES_IF_START + r"\1" + HELM_IF_END,
                    fdata,
                    re.M,
                )

                fdata = re.sub(
                    r"(.*command:\n)",
                    HELM_CONTROLER_DEP_POD_SEC_CTX_WITH_START +
                    HELM_ANNOTATIONS_TOYAML8 + HELM_IF_END + r"\1",
                    fdata,
                    re.M,
                )

            # make sure hostNetwork is not quoted as its a bool
            fdata = fdata.replace(
                "'{{ .Values.hostNetwork }}'", "{{ .Values.hostNetwork }}"
            )
            # make sure metrics.port is not quoted as its an int
            fdata = fdata.replace(
                "containerPort: '{{ .Values.metrics.port }}'", "containerPort: {{ .Values.metrics.port }}"
            )
            # make sure webhook is not quoted as its an int
            fdata = fdata.replace(
                "'{{ .Values.webhook.port }}'", "{{ .Values.webhook.port }}"
            )
            # make sure managerUserID is not quoted as its an int
            fdata = fdata.replace(
                "'{{ .Values.managerUserID }}'", "{{ .Values.managerUserID }}"
            )

            if not kind == "namespace":
                if (
                    "seldon1" in name
                    and name != "seldon1-manager-rolebinding"
                    and name != "seldon1-manager-role"
                    and name != "seldon1-manager-sas-role"
                    and name != "seldon1-manager-sas-rolebinding"
                ):
                    print("Ignore ", name)
                    continue
                elif (
                    name == "seldon-webhook-server-cert"
                    and kind == "secret"
                    or kind == "validatingwebhookconfiguration"
                ):
                    webhookData = webhookData + "---\n\n" + fdata
                else:
                    with open(filename, "w") as outfile:
                        outfile.write(fdata)
    # Write webhook related data in 1 file
    namespaceSelector = (
        "  namespaceSelector:\n    matchLabels:\n      seldon.io/controller-id: "
        + helm_namespace_override()
        + "\n"
    )
    objectSelector = (
        "  objectSelector:\n    matchLabels:\n      seldon.io/controller-id: "
        + helm_value("controllerId")
        + "\n"
    )
    kubeflowSelector = (
        "    matchLabels:\n      serving.kubeflow.org/inferenceservice: enabled\n"
    )
    webhookData = re.sub(
        r"(.*namespaceSelector:\n.*matchExpressions:\n.*\n.*\n)",
        HELM_VERSION_IF_START
        + HELM_NOT_SINGLE_NAMESPACE_IF_START
        + r"\1"
        + HELM_KUBEFLOW_IF_START
        + kubeflowSelector
        + HELM_IF_END
        + HELM_IF_END
        + HELM_IF_END
        + HELM_SINGLE_NAMESPACE_IF_START
        + namespaceSelector
        + HELM_IF_END,
        webhookData,
        re.M,
    )
    webhookData = re.sub(
        r"(.*objectSelector:\n.*matchExpressions:\n.*\n.*\n)",
        HELM_KUBEFLOW_IF_NOT_START
        + HELM_VERSION_IF_START
        + HELM_NOT_CONTROLLERID_IF_START
        + r"\1"
        + HELM_IF_END
        + HELM_IF_END
        + HELM_CONTROLLERID_IF_START
        + objectSelector
        + HELM_IF_END
        + HELM_IF_END,
        webhookData,
        re.M,
    )
    webhookData = webhookData + "\n" + HELM_IF_END

    print("Webhook data len", len(webhookData))

    filename = args.folder + "/" + "webhook.yaml"
    with open(filename, "w") as outfile:
        outfile.write(webhookData)
