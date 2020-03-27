import pytest
import logging

import seldon_core


logging.basicConfig(level=logging.DEBUG)


@pytest.fixture(params=[True, False])
def client_gets_metrics(monkeypatch, request):
    value = request.param
    monkeypatch.setattr(seldon_core.user_model, "CLIENT_GETS_METRICS", value)
    monkeypatch.setattr(seldon_core.seldon_methods, "CLIENT_GETS_METRICS", value)
    return value
