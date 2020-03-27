import pytest
import logging

import seldon_core


logging.basicConfig(level=logging.DEBUG)


@pytest.fixture(params=[True, False])
def client_gets_metrics(monkeypatch, request):
    value = request.param
    monkeypatch.setenv("INCLUDE_METRICS_IN_CLIENT_RESPONSE", str(value).lower())
    return value
