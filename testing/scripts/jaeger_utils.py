import requests
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_result

from seldon_e2e_utils import API_AMBASSADOR

JAEGER_QUERY_URL = f"http://{API_AMBASSADOR}/jaeger"


def _is_empty(result):
    return result is None or len(result) == 0


@retry(
    stop=stop_after_attempt(5),
    wait=wait_exponential(max=5),
    retry=retry_if_result(_is_empty),
)
def get_traces(pod_name, service, operation):
    endpoint = f"{JAEGER_QUERY_URL}/api/traces"
    params = {"service": service, "operation": operation, "tag": f"hostname:{pod_name}"}
    response = requests.get(endpoint, params=params)
    payload = response.json()
    traces = payload["data"]
    return traces
