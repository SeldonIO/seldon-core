import argparse
import glob

import yaml

parser = argparse.ArgumentParser()
parser.add_argument('--prefix', default="xx", help='find files matching prefix')
parser.add_argument('--folder', required=True, help='Output folder')
args, _ = parser.parse_known_args()

HELM_SPARTAKUS_IF_START = '{{- if .Values.usageMetrics.enabled }}\n'
HELM_RBAC_IF_START = '{{- if .Values.rbac.create }}\n'
HELM_SA_IF_START = '{{- if .Values.serviceAccount.create -}}\n'
HELM_IF_END = '{{- end }}\n'

HELM_ENV_SUBST = {
    "AMBASSADOR_ENABLED": "ambassador.enabled",
    "AMBASSADOR_SINGLE_NAMESPACE": "ambassador.singleNamespace",
    "ENGINE_SERVER_GRPC_PORT": "engine.grpc.port",
    "ENGINE_CONTAINER_IMAGE_PULL_POLICY": "engine.image.pullPolicy",
    "ENGINE_LOG_MESSAGES_EXTERNALLY": "engine.logMessagesExternally",
    "ENGINE_SERVER_PORT": "engine.port",
    "ENGINE_PROMETHEUS_PATH": "engine.prometheus.path",
    "ENGINE_CONTAINER_USER": "engine.user",
    "ENGINE_CONTAINER_SERVICE_ACCOUNT_NAME": "engine.serviceAccount.name",
    "ISTIO_ENABLED":"istio.enabled",
    "ISTIO_GATEWAY":"istio.gateway",
    "PREDICTIVE_UNIT_SERVICE_PORT":"predictiveUnit.port"

}
HELM_VALUES_IMAGE_PULL_POLICY = '{{ .Values.image.pullPolicy }}'


def helm_value(value: str):
    return '{{ .Values.' + value + ' }}'


if __name__ == "__main__":
    exp = args.prefix + "*"
    files = glob.glob(exp)
    for file in files:
        with open(file, 'r') as stream:
            res = yaml.safe_load(stream)
            kind = res["kind"].lower()
            name = res["metadata"]["name"].lower()
            filename = args.folder + "/" + (kind + "_" + name).lower() + ".yaml"

            if kind == "deployment" and name == "seldon-controller-manager":
                res["spec"]["template"]["spec"]["containers"][1]["imagePullPolicy"] = helm_value(
                    'image.pullPolicy')
                res["spec"]["template"]["spec"]["containers"][1][
                    "image"] = "{{ .Values.image.registry }}/{{ .Values.image.repository }}:{{ .Values.image.tag }}"

                for env in res["spec"]["template"]["spec"]["containers"][1]["env"]:
                    if env["name"] in HELM_ENV_SUBST:
                        env["value"] = helm_value(HELM_ENV_SUBST[env["name"]])
                    elif env["name"] == "ENGINE_CONTAINER_IMAGE_AND_VERSION":
                        env[
                            "value"] = '{{ .Values.engine.image.registry }}/{{ .Values.engine.image.repository }}:{{ .Values.engine.image.tag }}'

            if kind == "serviceaccount" and name == "seldon-manager":
                res["metadata"]["name"] = helm_value("serviceAccount.name")

            if kind == "clusterrolebinding" and name == "seldon-manager-rolebinding":
                res["subjects"][0]["name"] = helm_value("serviceAccount.name")

            fdata = yaml.dump(res, width=1000)

            # Spartatkus
            if name.find("spartakus") > -1:
                fdata = HELM_SPARTAKUS_IF_START + fdata + HELM_IF_END
            if name == "seldon-manager-rolebinding" or name == "seldon-manager-role":
                fdata = HELM_RBAC_IF_START + fdata + HELM_IF_END
            if name == "seldon-manager" and kind == "serviceaccount":
                fdata = HELM_SA_IF_START + fdata + HELM_IF_END

            with open(filename, 'w') as outfile:
                outfile.write(fdata)
