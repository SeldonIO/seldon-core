from seldon_e2e_utils import (initial_rest_request, retry_run,
                              to_resources_path, wait_for_rollout,
                              wait_for_status)


def test_xss_escaping(namespace):
    sdep_name = "mymodel"
    sdep_path = to_resources_path("graph-echo.json")
    retry_run(f"kubectl apply -f {sdep_path} -n {namespace}")
    wait_for_status(sdep_name, namespace)
    wait_for_rollout(sdep_name, namespace)

    payload = '<div class="div-class"></div>'

    # There is a small difference between the engine and the executor, where
    # the engine will escape the `=` symbol as its unicode equivalent, so we
    # need to consider both.
    expected = '\\u003cdiv class=\\"div-class\\"\\u003e\\u003c/div\\u003e'

    res = initial_rest_request(sdep_name, namespace, data=payload, dtype="strData")

    # We need to compare raw text (instead of `.json()`). Otherwise, Python
    # interprets the escaped sequences.
    assert expected in res.text


def test_xss_header(namespace):
    sdep_name = "mymodel"
    sdep_path = to_resources_path("graph-echo.json")
    retry_run(f"kubectl apply -f {sdep_path} -n {namespace}")
    wait_for_status(sdep_name, namespace)
    wait_for_rollout(sdep_name, namespace)

    res = initial_rest_request(sdep_name, namespace)

    assert "X-Content-Type-Options" in res.headers
    assert res.headers["X-Content-Type-Options"] == "nosniff"
