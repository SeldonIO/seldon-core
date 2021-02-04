import json
import subprocess
from subprocess import run


class TestBadGraphs(object):
    def test_duplicate_predictor_name(self):
        ret = run(
            "kubectl apply -f ../resources/bad_duplicate_predictor_name.json -n seldon",
            shell=True,
            check=False,
        )
        assert ret.returncode == 1

    # Name in graph and that in PodTemplateSpec don't match
    def test_model_name_mismatch(self):
        ret = run(
            "kubectl apply -f ../resources/bad_name_mismatch.json -n seldon",
            shell=True,
            check=False,
        )
        assert ret.returncode == 1

    # Name in graph and that in PodTemplateSpec don't match
    def test_model_no_graph(self):
        ret = run(
            "kubectl apply -f ../resources/bad_no_graph.yaml -n seldon",
            shell=True,
            check=False,
        )
        assert ret.returncode == 1
