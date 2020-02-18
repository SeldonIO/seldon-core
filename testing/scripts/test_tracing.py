from seldon_e2e_utils import (
    retry_run,
    wait_for_status,
    wait_for_rollout,
    initial_rest_request,
)
from jaeger_utils import get_traces


def test_tracing_rest(namespace):
    # Deploy model and check that is running
    retry_run(f"kubectl apply -f ../resources/graph-tracing.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    initial_rest_request("mymodel", namespace)

    traces = get_traces("executor", "predictions")
