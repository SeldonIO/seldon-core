from adserver.server import Protocol, CEServer, CEModel
from tornado.testing import AsyncHTTPTestCase
from typing import List, Dict
import json
import requests_mock


class TestProtocol(AsyncHTTPTestCase):
    def get_app(self):
        server = CEServer(Protocol.seldon_http, "a,b,c", "x.y.z")
        return server.create_application()

    def test_seldon_protocol(self):
        response = self.fetch("/protocol")
        self.assertEqual(response.code, 200)
        self.assertEqual(response.body.decode("utf-8"), str(Protocol.seldon_http.value))


customHeaderKey = "Seldonheader"
customHeaderVal = "SeldonValue"


class DummyModel(CEModel):
    def __init__(self, name: str, create_response: bool = True):
        super().__init__(name)
        self.create_response = create_response

    @staticmethod
    def getResponse() -> Dict:
        return {"foo": 1}

    def load(self):
        pass

    def process_event(self, inputs: List, headers: Dict) -> Dict:
        assert headers[customHeaderKey] == customHeaderVal
        if self.create_response:
            return DummyModel.getResponse()


class TestSeldonHttpModel(AsyncHTTPTestCase):
    def setupEnv(self):
        self.replyUrl = "http://reply-location"
        self.eventSource = "x.y.z"
        self.eventType = "a.b.c"

    def get_app(self):
        self.setupEnv()
        server = CEServer(
            Protocol.seldon_http, self.eventType, self.eventSource, 9000, self.replyUrl
        )
        model = DummyModel("name")
        server.register_model(model)
        return server.create_application()

    def test_basic(self):
        data = {"data": {"ndarray": [[1, 2, 3]]}}
        dataStr = json.dumps(data)
        with requests_mock.Mocker() as m:
            m.post(self.replyUrl, text="resp")

            response = self.fetch(
                "/",
                method="POST",
                body=dataStr,
                headers={
                    customHeaderKey: customHeaderVal,
                    "ce-source": "a.b.c",
                    "ce-type": "d.e.f",
                    "ce-id": "1234",
                    "ce-specversion": "1.0",
                },
            )
            self.assertEqual(response.code, 200)
            expectedResponse = json.dumps(DummyModel.getResponse())
            self.assertEqual(response.body.decode("utf-8"), expectedResponse)
            self.assertEqual(
                m.request_history[0].json(), json.dumps(DummyModel.getResponse())
            )
            headers: Dict = m.request_history[0]._request.headers
            self.assertEqual(headers["ce-source"], self.eventSource)
            self.assertEqual(headers["ce-type"], self.eventType)


class TestKFservingV2HttpModel(AsyncHTTPTestCase):
    def setupEnv(self):
        self.replyUrl = "http://reply-location"
        self.eventSource = "x.y.z"
        self.eventType = "a.b.c"

    def get_app(self):
        self.setupEnv()
        server = CEServer(
            Protocol.kfserving_http,
            self.eventType,
            self.eventSource,
            9000,
            self.replyUrl,
        )
        model = DummyModel("name")
        server.register_model(model)
        return server.create_application()

    def test_basic(self):
        data = {
            "inputs": [
                {
                    "name": "input_1",
                    "datatype": "FP32",
                    "shape": [1, 3],
                    "data": [1, 2, 3],
                }
            ]
        }
        dataStr = json.dumps(data)
        with requests_mock.Mocker() as m:
            m.post(self.replyUrl, text="resp")

            response = self.fetch(
                "/",
                method="POST",
                body=dataStr,
                headers={
                    customHeaderKey: customHeaderVal,
                    "ce-source": "a.b.c",
                    "ce-type": "d.e.f",
                    "ce-id": "1234",
                    "ce-specversion": "1.0",
                },
            )
            self.assertEqual(response.code, 200)
            expectedResponse = json.dumps(DummyModel.getResponse())
            self.assertEqual(response.body.decode("utf-8"), expectedResponse)
            self.assertEqual(
                m.request_history[0].json(), json.dumps(DummyModel.getResponse())
            )
            headers: Dict = m.request_history[0]._request.headers
            self.assertEqual(headers["ce-source"], self.eventSource)
            self.assertEqual(headers["ce-type"], self.eventType)


class TestModelNoResponse(AsyncHTTPTestCase):
    def setupEnv(self):
        self.replyUrl = "http://reply-location"
        self.eventSource = "x.y.z"
        self.eventType = "a.b.c"

    def get_app(self):
        self.setupEnv()
        server = CEServer(
            Protocol.seldon_http, self.eventType, self.eventSource, 9000, self.replyUrl
        )
        model = DummyModel("name", create_response=False)
        server.register_model(model)
        return server.create_application()

    def test_basic(self):
        data = {"data": {"ndarray": [[1, 2, 3]]}}
        dataStr = json.dumps(data)
        with requests_mock.Mocker() as m:
            m.post(self.replyUrl, text="resp")
            response = self.fetch(
                "/",
                method="POST",
                body=dataStr,
                headers={
                    customHeaderKey: customHeaderVal,
                    "ce-source": "a.b.c",
                    "ce-type": "d.e.f",
                    "ce-id": "1234",
                    "ce-specversion": "1.0",
                },
            )
            self.assertEqual(response.code, 200)
            self.assertEqual(response.body, b"")
            self.assertEqual(len(m.request_history), 0)
