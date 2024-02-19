from adserver.server import Protocol, CEServer, CEModel
from tornado.testing import AsyncHTTPTestCase
from adserver.base import ModelResponse
from typing import List, Dict, Optional, Union
import json
import requests_mock
from cloudevents.sdk import converters
from cloudevents.sdk import marshaller
from cloudevents.sdk.event import v1


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
    def getResponse() -> ModelResponse:
        return ModelResponse(data={"foo": 1}, metrics=None)

    def load(self):
        pass

    def process_event(self, inputs: Union[List, Dict], headers: Dict):
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
            expectedResponse = DummyModel.getResponse().data
            # assert that the expected response conforms to the CloudEvent spec
            event = v1.Event()
            http_marshaller = marshaller.NewDefaultHTTPMarshaller()
            try:
                event = http_marshaller.FromRequest(
                    event, response.headers, response.body, json.loads
                )
            except Exception as e:
                assert False, f"Failed to unmarshall data with error: {type(e).__name__}('{e}')"

            # assert cloud event properties have been set correctly in response
            self.assertEqual(event.Data(), expectedResponse)
            self.assertEqual(event.Source(), self.eventSource)
            self.assertEqual(event.EventType(), self.eventType)
            self.assertEqual(event.ContentType(), "application/json")
            self.assertEqual(event.EventID(), "1234")
            self.assertEqual(event.CloudEventVersion(), "1.0")
            self.assertEqual(response.body.decode("utf-8"), json.dumps(expectedResponse))

            # assert requests have been made with the correct headers and data
            self.assertEqual(m.request_history[0].json(), expectedResponse)
            headers: Dict = m.request_history[0]._request.headers
            self.assertEqual(headers["ce-source"], self.eventSource)
            self.assertEqual(headers["ce-type"], self.eventType)
            self.assertNotIn("ce-datacontenttype", headers)


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
            expectedResponse = DummyModel.getResponse().data
            self.assertEqual(response.body.decode("utf-8"), json.dumps(expectedResponse))
            self.assertEqual(m.request_history[0].json(), expectedResponse)
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
