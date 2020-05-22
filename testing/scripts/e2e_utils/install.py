import os

from sh import helm, kubectl

SC_ROOT_PATH = os.path.dirname(os.path.dirname(os.path.dirname(__file__),))
HELM_CHARTS_PATH = os.path.dirname(SC_ROOT_PATH, "helm-charts")

SC_NAME = "seldon"
SC_NAMESPACE = "seldon-system"


def install_seldon(name=SC_NAME, namespace=SC_NAMESPACE, executor=True, version=None):
    chart_path = "seldonio/seldon-core-operator"
    if version is None:
        # Use local
        chart_path = os.path.join(HELM_CHARTS_PATH, "seldon-core-operator")

    values = [
        "istio.enabled=true",
        "istio.gateway=istio-system/seldon-gateway",
        "certManager.enabled=false",
    ]

    if not executor:
        values.push("executor.enabled=false")

    helm.install(
        name, chart_path, namespace=namespace, set=values, version=version, wait=True,
    )


def delete_seldon(name=SC_NAME, namespace=SC_NAMESPACE):
    helm.delete(name, namespace=namespace)

    # Helm 3.0.3 doesn't delete CRDs
    kubectl.delete(
        "crd", "seldondeployments.machinelearning.seldon.io", ignore_not_found=True,
    )
