import requests
from tenacity import retry, retry_if_result, stop_after_attempt, wait_exponential

from seldon_e2e_utils import API_AMBASSADOR


def _is_transient_error(res):
    to_retry = [404, 503, 504]
    return res.status_code in to_retry


@retry(
    wait=wait_exponential(multiplier=1),
    stop=stop_after_attempt(3),
    retry=retry_if_result(_is_transient_error),
)
def inference_request(
    deployment_name: str,
    model_name: str,
    namespace: str,
    payload: dict,
    host: str = API_AMBASSADOR,
):
    endpoint = f"http://{host}/seldon/{namespace}/{deployment_name}/v2/models/{model_name}/infer"
    return requests.post(endpoint, json=payload)
