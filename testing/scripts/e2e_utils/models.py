import os

from sh import helm

from .common import HELM_CHARTS_PATH, to_helm_values_list


def deploy_model(model_name: str, namespace: str, **kwargs):
    chart_path = os.path.join(HELM_CHARTS_PATH, "seldon-single-model")

    # Convert "_" to "." on kwargs
    values = {key.replace("_", "."): val for key, val in kwargs.items()}

    helm.install(
        model_name,
        chart_path,
        to_helm_values_list(values),
        namespace=namespace,
        wait=True,
    )


def delete_model(model_name: str, namespace: str):
    helm.delete(model_name, namespace=namespace)
