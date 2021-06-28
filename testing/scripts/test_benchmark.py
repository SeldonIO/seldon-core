import pytest
import json
import logging

import pandas as pd

from seldon_e2e_utils import post_comment_in_pr, run_benchmark_and_capture_results


@pytest.mark.benchmark
@pytest.mark.usefixtures("argo_worfklows")
def test_service_orchestrator():

    df = run_benchmark_and_capture_results(
            api_type_list=["rest", "grpc"],
            disable_orchestrator_list=["false", "true"],
        )

    result_body = "# Benchmark results\n\n"
    result_body += str(df.to_markdown())
    post_comment_in_pr(result_body)

    # Ensure all mean performance latency below 5 ms
    assert all(df["mean"] < 5)
    # Ensure 99th percentiles are not spiking above 15ms
    assert all(df["mean"] < 15)
    # Ensure throughput is above 200 rps for REST
    assert all(df[df["apiType"] == "rest"]["throughputAchieved"] > 200)
    # Ensure throughput is above 250 rps for GRPC
    assert all(df[df["apiType"] == "grpc"]["throughputAchieved"] > 250)
    # Validate latenc added by adding service orchestrator is lower than 4ms
    assert all((df[df["disableOrchestrator"] == "true"]["mean"].values
            - df[df["disableOrchestrator"] == "false"]["mean"].values) < 4)
    assert all((df[df["disableOrchestrator"] == "true"]["99th"].values
            - df[df["disableOrchestrator"] == "false"]["99th"].values) < 4)

