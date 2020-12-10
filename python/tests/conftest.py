import os
import pytest
import logging
import seldon_core

from seldon_core.wrapper import get_rest_microservice
from seldon_core.metrics import SeldonMetrics
from .helpers import UserObject, MicroserviceWrapper

logging.basicConfig(level=logging.DEBUG)

RESOURCES_PATH = os.path.join(os.path.dirname(__file__), "resources")


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


@pytest.fixture
def microservice(request):
    # Make it compatible for both direct and indirect usage
    opts = {}
    if hasattr(request, "param"):
        opts = request.param

    # Extract opts from request' param
    app_name = opts.get("app_name", "model-template-app")
    app_location = opts.get("app_location", os.path.join(RESOURCES_PATH, app_name))
    envs = opts.get("envs", {})
    tracing = opts.get("tracing", False)

    wrapper = MicroserviceWrapper(app_location=app_location, envs=envs, tracing=tracing)

    with wrapper:
        yield wrapper


@pytest.fixture
def app():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    return get_rest_microservice(user_object, seldon_metrics)
