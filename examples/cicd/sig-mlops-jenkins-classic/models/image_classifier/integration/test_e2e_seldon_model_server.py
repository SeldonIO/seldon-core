from seldon_core.seldon_client import SeldonClient
from seldon_core.utils import seldon_message_to_json
import numpy as np
from subprocess import run
import time
import logging


API_AMBASSADOR = "localhost:8003"


def test_sklearn_server(data):
    y_true = [5, 0]

    sc = SeldonClient(
        gateway="ambassador",
        gateway_endpoint=API_AMBASSADOR,
        deployment_name="image-classifier",
        payload_type="ndarray",
        namespace="seldon",
        transport="rest",
    )

    sm_result = sc.predict(data=np.array(data))
    logging.info(sm_result)
    result = seldon_message_to_json(sm_result.response)
    values = result.get("data", {}).get("ndarray", {})
    assert values == labels
