from __future__ import absolute_import, division, print_function

from contextlib import contextmanager
import json
import os
from os.path import dirname, join
import socket
from subprocess import Popen
import time
import requests


@contextmanager
def start_microservice(app_location):
    p = None
    try:
        # PYTHONUNBUFFERED=x
        # exec python -u microservice.py $MODEL_NAME $API_TYPE --service-type $SERVICE_TYPE --persistence $PERSISTENCE
        env_vars = dict(os.environ)
        env_vars.update({
            "PYTHONUNBUFFERED": "x",
            "PYTHONPATH": app_location,
            "APP_HOST": "127.0.0.1",
            "SERVICE_PORT_ENV_NAME": "5000",
        })
        with open(join(app_location, ".s2i", "environment")) as fh:
            for line in fh.readlines():
                line = line.strip()
                if line:
                    key, value = line.split("=", 1)
                    key, value = key.strip(), value.strip()
                    if key and value:
                        env_vars[key] = value
        cmd = (
            "seldon-core-microservice",
            env_vars["MODEL_NAME"],
            env_vars["API_TYPE"],
            "--service-type", env_vars["SERVICE_TYPE"],
            "--persistence", env_vars["PERSISTENCE"],
        )
        print("starting:", " ".join(cmd))
        print("cwd:", app_location)
        # stdout=PIPE, stderr=PIPE,
        p = Popen(cmd, cwd=app_location, env=env_vars,)

        for q in range(10):
            time.sleep(0.1)
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            result = sock.connect_ex(("127.0.0.1", 5000))
            if result == 0:
                break
        else:
            raise RuntimeError("Server did not bind to 127.0.0.1:5000")
        yield
    finally:
        if p:
            p.terminate()


def test_model_template_app():
    with start_microservice(join(dirname(__file__), "model-template-app")):
        data = '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
        response = requests.get(
            "http://127.0.0.1:5000/predict", params="json=%s" % data)
        response.raise_for_status()
        assert response.json() == {
            'data': {'names': ['t:0', 't:1'], 'ndarray': [[1.0, 2.0]]}, 'meta': {}}

        data = ('{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},'
                '"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],'
                '"ndarray":[[1.0,2.0]]}},"reward":1}')
        response = requests.get(
            "http://127.0.0.1:5000/send-feedback", params="json=%s" % data)
        response.raise_for_status()
        assert response.json() == {}


def test_tester_model_template_app():
    # python api-tester.py contract.json  0.0.0.0 8003 --oauth-key oauth-key --oauth-secret oauth-secret -p --grpc --oauth-port 8002 --endpoint send-feedback
    # python tester.py contract.json 0.0.0.0 5000 -p --grpc
    with start_microservice(join(dirname(__file__), "model-template-app")):
        env_vars = dict(os.environ)
        cmd = (
            "seldon-core-tester",
            join(dirname(__file__), "model-template-app", "contract.json"),
            "127.0.0.1",
            "5000",
            "--prnt",
        )
        print("starting:", " ".join(cmd))
        p = Popen(cmd, env=env_vars,)  # stdout=PIPE, stderr=PIPE,
        p.wait()
        assert p.returncode == 0

    """
    starting: seldon-core-tester tests/model-template-app/contract.json 127.0.0.1 5000 --prnt
    ----------------------------------------
    SENDING NEW REQUEST:
    {'meta': {}, 'data': {'names': ['sepal_length', 'sepal_width', 'petal_length', 'petal_width'], 'ndarray': [[5.627, 2.239, 9.407, 2.604]]}}
    RECEIVED RESPONSE:
    {'data': {'names': ['t:0', 't:1', 't:2', 't:3'], 'ndarray': [[5.627, 2.239, 9.407, 2.604]]}}

    Time 0.010219097137451172
    """
