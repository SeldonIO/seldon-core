import requests

from seldon_e2e_utils import API_ISTIO_GATEWAY


def inference_request(
    model_name: str, namespace: str, payload: dict, host: str = API_ISTIO_GATEWAY
):
    endpoint = (
        f"http://{host}/seldon/{namespace}/{model_name}/v2/models/{model_name}/infer"
    )
    return requests.post(endpoint, json=payload)
