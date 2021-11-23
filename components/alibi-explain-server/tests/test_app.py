import json

from tornado.testing import AsyncHTTPTestCase

from alibiexplainer import AlibiExplainer, Protocol
from alibiexplainer.server import ExplainerServer
from alibiexplainer.utils import ExplainerMethod

from .make_test_models import make_anchor_tabular
from .utils import SKLearnServer

IRIS_MODEL_URI = "gs://seldon-models/v1.11.0-dev/sklearn/iris/*"


class TestExplainerApp(AsyncHTTPTestCase):
    def get_app(self):
        app_server = ExplainerServer()

        # predict_fn
        skmodel = SKLearnServer(IRIS_MODEL_URI)
        skmodel.load()

        # just an arbitrary explainer
        alibi_model = make_anchor_tabular()

        # create wrapper
        alibi_model_wrapper = AlibiExplainer(
            name="dummy",
            predict_fn=skmodel.predict,
            method=ExplainerMethod.anchor_tabular,
            config={},
            explainer=alibi_model,
            protocol=Protocol.seldon_http,
        )

        app_server.register_model(alibi_model_wrapper)

        return app_server.create_application()

    def test_explain(self):
        data = '{"data": {"ndarray": [[0.1, 0.1, 0.1, 0.1]]}}'
        for endpoint in [
            "/api/v1.0/explain",
            "/api/v0.1/explain",
            "/v1/models/dummy:explain",
        ]:
            response = self.fetch(path=endpoint, method="POST", body=data)
            exp_json = json.loads(response.body)
            assert exp_json["meta"]["name"] == "AnchorTabular"
