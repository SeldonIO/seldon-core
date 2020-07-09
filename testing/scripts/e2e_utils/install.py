import os

from sh import helm, kubectl

SC_ROOT_PATH = os.path.abspath(
    os.path.join(
        __file__, os.path.pardir, os.path.pardir, os.path.pardir, os.path.pardir
    )
)
HELM_CHARTS_PATH = os.path.join(SC_ROOT_PATH, "helm-charts")

SC_NAME = "seldon"
SC_NAMESPACE = "seldon-system"


def install_seldon(name=SC_NAME, namespace=SC_NAMESPACE, executor=True, version=None):
    chart_path = "seldonio/seldon-core-operator"
    if version is None:
        # Use local
        chart_path = os.path.join(HELM_CHARTS_PATH, "seldon-core-operator")

    values = {
        "istio.enabled": "true",
        "istio.gateway": "istio-system/seldon-gateway",
        "certManager.enabled": "false",
    }

    if not executor:
        values["executor.enabled"] = "false"

    helm.install(
        name,
        chart_path,
        _to_helm_values_list(values),
        namespace=namespace,
        version=version,
        wait=True,
    )


def delete_seldon(name=SC_NAME, namespace=SC_NAMESPACE):
    helm.delete(name, namespace=namespace)

    # Helm 3.0.3 doesn't delete CRDs
    kubectl.delete(
        "crd", "seldondeployments.machinelearning.seldon.io", ignore_not_found=True
    )


def _to_helm_values_list(values):
    """
    The sh lib doesn't allow you to specify multiple instances of the same
    kwarg. https://github.com/amoffat/sh/issues/529

    The best option is to concatenate them into a list.
    """
    values_list = []
    for key, val in values.items():
        values_list += ["--set", f"{key}={val}"]

    return values_list
