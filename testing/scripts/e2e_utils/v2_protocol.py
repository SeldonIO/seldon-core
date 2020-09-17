import requests

from seldon_e2e_utils import API_AMBASSADOR


def inference_request(
    model_name: str, namespace: str, payload: dict, host: str = API_AMBASSADOR
):
    endpoint = (
        f"http://{host}/seldon/{namespace}/{model_name}/v2/models/{model_name}/infer"
    )
    return requests.post(endpoint, json=payload)
