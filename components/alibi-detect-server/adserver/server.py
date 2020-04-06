import json
import logging
import os
import uuid
from http import HTTPStatus
from typing import Dict, Optional

import requests
import tornado.httpserver
import tornado.ioloop
import tornado.web
from adserver.base import CEModel
from adserver.protocols.request_handler import RequestHandler
from adserver.protocols.seldon_http import SeldonRequestHandler
from adserver.protocols.tensorflow_http import TensorflowRequestHandler
from cloudevents.sdk import converters
from cloudevents.sdk import marshaller
from cloudevents.sdk.event import v02
from adserver.protocols import Protocol

DEFAULT_HTTP_PORT = 8080
CESERVER_LOGLEVEL = os.environ.get("CESERVER_LOGLEVEL", "INFO").upper()
logging.basicConfig(level=CESERVER_LOGLEVEL)


class CEServer(object):
    def __init__(
        self,
        protocol: Protocol,
        event_type: str,
        event_source: str,
        http_port: int = DEFAULT_HTTP_PORT,
        reply_url: str = None,
    ):
        """
        CloudEvents server

        Parameters
        ----------
        protocol
             wire protocol
        http_port
             http port to listen on
        reply_url
             reply url to send response event
        event_type
             type of event being handled (for req logging purposes)
        """
        self.registered_model: Optional[CEModel] = None
        self.http_port = http_port
        self.protocol = protocol
        self.reply_url = reply_url
        self._http_server: Optional[tornado.httpserver.HTTPServer] = None
        self.event_type = event_type
        self.event_source = event_source

    def create_application(self):
        return tornado.web.Application(
            [
                # Outlier detector
                (
                    r"/",
                    EventHandler,
                    dict(
                        protocol=self.protocol,
                        model=self.registered_model,
                        reply_url=self.reply_url,
                        event_type=self.event_type,
                        event_source=self.event_source,
                    ),
                ),
                # Protocol Discovery API that returns the serving protocol supported by this server.
                (r"/protocol", ProtocolHandler, dict(protocol=self.protocol)),
                # Prometheus Metrics API that returns metrics for model servers
                (r"/v1/metrics", MetricsHandler, dict(model=self.registered_model)),
            ]
        )

    def start(self, model: CEModel):
        """
        Start the server

        Parameters
        ----------
        model
             The model to load

        """
        self.register_model(model)

        self._http_server = tornado.httpserver.HTTPServer(self.create_application())

        logging.info("Listening on port %s" % self.http_port)
        self._http_server.bind(self.http_port)
        self._http_server.start(1)  # Single worker at present
        tornado.ioloop.IOLoop.current().start()

    def register_model(self, model: CEModel):
        if not model.name:
            raise Exception("Failed to register model, model.name must be provided.")
        self.registered_model = model
        logging.info("Registering model:" + model.name)


def get_request_handler(protocol, request: Dict) -> RequestHandler:
    """
    Create a request handler for the data

    Parameters
    ----------
    protocol
         Protocol to use
    request
         The incoming request
    Returns
    -------
         A Request Handler for the desired protocol

    """
    if protocol == Protocol.tensorflow_http:
        return TensorflowRequestHandler(request)
    else:
        return SeldonRequestHandler(request)


def sendCloudEvent(event: v02.Event, url: str):
    """
    Send CloudEvent

    Parameters
    ----------
    event
         CloudEvent to send
    url
         Url to send event

    """
    http_marshaller = marshaller.NewDefaultHTTPMarshaller()
    binary_headers, binary_data = http_marshaller.ToRequest(
        event, converters.TypeBinary, json.dumps
    )

    print("binary CloudEvent")
    for k, v in binary_headers.items():
        print("{0}: {1}\r\n".format(k, v))
    print(binary_data)

    response = requests.post(url, headers=binary_headers, data=binary_data)
    response.raise_for_status()


class EventHandler(tornado.web.RequestHandler):
    def initialize(
        self,
        protocol: str,
        model: CEModel,
        reply_url: str,
        event_type: str,
        event_source: str,
    ):
        """
        Event Handler

        Parameters
        ----------
        protocol
             The protocol to expect
        model
             The model to use
        reply_url
             The reply url to send model responses
        event_type
             The CE event type to be sent
        event_source
             The CE event source
        """
        self.protocol = protocol
        self.model = model
        self.reply_url = reply_url
        self.event_type = event_type
        self.event_source = event_source

    def post(self):
        """
        Handle post request. Extract data. Call event handler and optionally send a reply event.

        """
        if not self.model.ready:
            self.model.load()

        try:
            body = json.loads(self.request.body)
        except json.decoder.JSONDecodeError as e:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason="Unrecognized request format: %s" % e,
            )

        # Extract payload from request
        request_handler: RequestHandler = get_request_handler(self.protocol, body)
        request_handler.validate()
        request = request_handler.extract_request()

        # Create event from request body
        event = v02.Event()
        http_marshaller = marshaller.NewDefaultHTTPMarshaller()
        event = http_marshaller.FromRequest(
            event, self.request.headers, self.request.body, json.loads
        )
        logging.debug(json.dumps(event.Properties()))

        # Extract any desired request headers
        headers = {}

        for (key, val) in self.request.headers.get_all():
            headers[key] = val

        response = self.model.process_event(request, headers)

        if response is not None:
            responseStr = json.dumps(response)

            # Create event from response if reply_url is active
            if not self.reply_url == "":
                if event.EventID() is None or event.EventID() == "":
                    resp_event_id = uuid.uuid1().hex
                else:
                    resp_event_id = event.EventID()
                revent = (
                    v02.Event()
                    .SetContentType("application/json")
                    .SetData(responseStr)
                    .SetEventID(resp_event_id)
                    .SetSource(self.event_source)
                    .SetEventType(self.event_type)
                    .SetExtensions(event.Extensions())
                )
                logging.debug(json.dumps(revent.Properties()))
                sendCloudEvent(revent, self.reply_url)

            self.write(json.dumps(response))


class LivenessHandler(tornado.web.RequestHandler):
    def get(self):
        self.write("Alive")


class ProtocolHandler(tornado.web.RequestHandler):
    def initialize(self, protocol: Protocol):
        self.protocol = protocol

    def get(self):
        self.write(str(self.protocol.value))


class MetricsHandler(tornado.web.RequestHandler):
    def initialize(self, model: CEModel):
        self.model = model

    def get(self):
        self.write("Not Implemented")
