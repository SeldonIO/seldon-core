from seldon_e2e_utils import (
    get_pod_names,
    get_deployment_names,
    retry_run,
    wait_for_status,
    wait_for_rollout,
    initial_rest_request,
)
from jaeger_utils import get_traces


def assert_trace(trace, expected_operations):
    spans = trace["spans"]
    assert len(spans) == len(expected_operations)

    # Assert the operations are in the right order
    for idx, span in enumerate(spans):
        assert span["operationName"] == expected_operations[idx]
        if idx > 0:
            # Assert there is only one ref to the previous span
            refs = span["references"]
            assert len(refs) == 1

            prev_span = spans[idx - 1]
            ref = span["references"][0]
            assert ref["refType"] == "CHILD_OF"
            assert ref["spanID"] == prev_span["spanID"]


def test_tracing_rest(namespace):
    # Deploy model and check that is running
    retry_run(f"kubectl apply -f ../resources/graph-tracing.json -n {namespace}")
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)
    initial_rest_request("mymodel", namespace)

    # We need the current pod name to find the right traces
    deployment_names = get_deployment_names("mymodel", namespace)
    deployment_name = deployment_names[0]
    pod_names = get_pod_names(deployment_name, namespace)
    pod_name = pod_names[0]

    # Get traces and assert their content
    traces = get_traces(pod_name, "executor", "predictions")
    assert len(traces) == 1

    trace = traces[0]
    processes = trace["processes"]
    assert len(processes) == 2
    assert_trace(trace, expected_operations=["predictions", "/predict", "Predict"])
