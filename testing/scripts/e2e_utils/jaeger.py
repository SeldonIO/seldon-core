import requests
from seldon_e2e_utils import API_AMBASSADOR
from tenacity import (retry, retry_if_result, stop_after_attempt,
                      wait_exponential)

JAEGER_QUERY_URL = f"http://{API_AMBASSADOR}/jaeger"


def _is_empty(result):
    return result is None or len(result) == 0


@retry(
    stop=stop_after_attempt(5),
    wait=wait_exponential(max=5),
    retry=retry_if_result(_is_empty),
)
def get_traces(pod_name, service, operation, _should_retry=lambda x: False):
    """
    Fetch traces for a given pod, service and operation.

    We use Jaeger's [**internal** REST
    API](https://www.jaegertracing.io/docs/1.13/apis/#http-json-internal).
    Therefore, it may stop working at some point!

    Note that this method will get retried 5 times (with an exponentially
    growing waiting time) if the traces are empty.
    This is to give time to Jaeger to collect and process the trace, which is
    performed asynchronously.

    Parameters
    ---
    pod_name : str
        We currently don't have access to the PUID (see
        https://github.com/SeldonIO/seldon-core/issues/1460).
        As a workaround, we filter the traces using the Pod name.
    service : str
        Service sending the traces.
        This will usually be the `'executor'`, since it's the one which creates
        the trace.
    operation : str
        Operation which was traced (e.g. `'predictions'`).

    Returns
    ---
    traces : list
        List of traces, where each trace contains spans, processes, etc.
    """
    endpoint = f"{JAEGER_QUERY_URL}/api/traces"
    params = {"service": service, "operation": operation, "tag": f"hostname:{pod_name}"}
    response = requests.get(endpoint, params=params)
    payload = response.json()
    traces = payload["data"]

    if _should_retry(traces):
        return None

    return traces
