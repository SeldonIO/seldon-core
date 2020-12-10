import requests

from tenacity import retry, wait_exponential, stop_after_attempt, retry_if_result
from seldon_e2e_utils import API_AMBASSADOR


def _is_404(res):
    return res.status_code == 404


@retry(
    wait=wait_exponential(multiplier=1),
    stop=stop_after_attempt(3),
    retry=retry_if_result(_is_404),
)
def inference_request(
    model_name: str, namespace: str, payload: dict, host: str = API_AMBASSADOR
):
    endpoint = (
        f"http://{host}/seldon/{namespace}/{model_name}/v2/models/{model_name}/infer"
    )
    return requests.post(endpoint, json=payload)
