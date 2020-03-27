import pytest
import logging

import seldon_core


logging.basicConfig(level=logging.DEBUG)


@pytest.fixture(params=[True, False])
def client_gets_metrics(monkeypatch, request):
    value = request.param
    monkeypatch.setattr(
        seldon_core.user_model, "INCLUDE_METRICS_IN_CLIENT_RESPONSE", value
    )
    monkeypatch.setattr(
        seldon_core.seldon_methods, "INCLUDE_METRICS_IN_CLIENT_RESPONSE", value
    )
    return value
