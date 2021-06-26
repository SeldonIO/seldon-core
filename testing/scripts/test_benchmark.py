import pytest

from seldon_e2e_utils import create_and_run_script


@pytest.mark.benchmark
def test_benchmark_overall(namespace):
    create_and_run_script("../../examples/batch/benchmarking-argo-workflows", "README")
