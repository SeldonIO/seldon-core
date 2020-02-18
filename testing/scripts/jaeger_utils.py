import requests

from seldon_e2e_utils import API_AMBASSADOR

JAEGER_QUERY_URL = f"http://{API_AMBASSADOR}/jaeger"


def get_traces(service, operation):
    endpoint = f"{JAEGER_QUERY_URL}/api/traces"
    params = dict(service=service, operation=operation)
    response = requests.get(endpoint, params=params)
    payload = response.json()
    return payload["data"]
