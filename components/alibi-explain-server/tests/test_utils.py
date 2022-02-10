import pytest
import numpy as np

from requests_mock import Mocker

from alibiexplainer.utils import (
    construct_predict_fn,
    Protocol,
    SELDON_SKIP_LOGGING_HEADER,
    SELDON_PREDICTOR_URL_FORMAT,
    TENSORFLOW_PREDICTOR_URL_FORMAT,
)


@pytest.mark.parametrize("protocol", [Protocol.seldon_http, Protocol.tensorflow_http])
def test_construct_predict_fn(protocol: str, requests_mock: Mocker):
    predictor_host = "fake-endpoint.com"
    model_name = "foo"
    predictor_endpoint = SELDON_PREDICTOR_URL_FORMAT.format(predictor_host)
    if protocol == Protocol.tensorflow_http:
        predictor_endpoint = TENSORFLOW_PREDICTOR_URL_FORMAT.format(
            predictor_host, model_name
        )

    res_value = [[7]]
    requests_mock.post(
        predictor_endpoint,
        json={"data": {"ndarray": res_value}, "predictions": res_value},
    )

    predict_fn = construct_predict_fn(
        predictor_host, model_name=model_name, protocol=protocol
    )

    res = predict_fn(arr=[[0, 1, 2, 3]])

    assert res == np.array(res_value)
    assert requests_mock.call_count == 1
    assert SELDON_SKIP_LOGGING_HEADER in requests_mock.last_request.headers
    assert requests_mock.last_request.headers[SELDON_SKIP_LOGGING_HEADER] == "true"
