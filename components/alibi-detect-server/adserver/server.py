import json
import logging
import os
from http import HTTPStatus
from typing import Dict, Optional

import requests
import tornado.httpserver
import tornado.ioloop
import tornado.web
from adserver.base import CEModel, ModelResponse
from adserver.protocols.request_handler import RequestHandler
from adserver.protocols.seldon_http import SeldonRequestHandler
from adserver.protocols.seldonfeedback_http import SeldonFeedbackRequestHandler
from adserver.protocols.tensorflow_http import TensorflowRequestHandler
from adserver.protocols.v2 import KFservingV2RequestHandler
from cloudevents.sdk import converters
from cloudevents.sdk import marshaller
from cloudevents.sdk.event import v1
from adserver.protocols import Protocol
from seldon_core.metrics import SeldonMetrics, validate_metrics
import uuid

DEFAULT_HTTP_PORT = 8080
CESERVER_LOGLEVEL = os.environ.get("CESERVER_LOGLEVEL", "INFO").upper()
logging.basicConfig(level=CESERVER_LOGLEVEL)

DEFAULT_LABELS = {
    "seldon_deployment_namespace": os.environ.get(
        "SELDON_DEPLOYMENT_NAMESPACE", "NOT_IMPLEMENTED"
    )
}


class CEServer(object):
    def __init__(
        self,
        protocol: Protocol,
        event_type: str,
        event_source: str,
        http_port: int = DEFAULT_HTTP_PORT,
        reply_url: Optional[str] = None,
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
        self.seldon_metrics = SeldonMetrics(
            extra_default_labels=DEFAULT_LABELS,
        )

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
                        seldon_metrics=self.seldon_metrics,
                    ),
                ),
                # Protocol Discovery API that returns the serving protocol supported by this server.
                (r"/protocol", ProtocolHandler, dict(protocol=self.protocol)),
                # Prometheus Metrics API that returns metrics for model servers
                (
                    r"/v1/metrics",
                    MetricsHandler,
                    dict(seldon_metrics=self.seldon_metrics),
                ),
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
    elif protocol == Protocol.seldon_http:
        return SeldonRequestHandler(request)
    elif protocol == Protocol.seldonfeedback_http:
        return SeldonFeedbackRequestHandler(request)
    elif protocol == Protocol.kfserving_http:
        return KFservingV2RequestHandler(request)
    else:
        raise Exception(f"Unknown protocol {protocol}")


def forward_request(headers, data, url):
    """
    Forward request

    Parameters
    ----------
    headers
         Headers to forward
    data
         Data to forward
    url
         Url to forward to

    """
    response = requests.post(url, headers=headers, data=data)
    response.raise_for_status()


class EventHandler(tornado.web.RequestHandler):
    def initialize(
        self,
        protocol: str,
        model: CEModel,
        reply_url: str,
        event_type: str,
        event_source: str,
        seldon_metrics: SeldonMetrics,
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
        self.seldon_metrics = seldon_metrics

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
        event = v1.Event()
        http_marshaller = marshaller.NewDefaultHTTPMarshaller()
        event = http_marshaller.FromRequest(
            event, self.request.headers, self.request.body, json.loads
        )
        logging.debug(json.dumps(event.Properties()))

        # Extract any desired request headers
        headers = {}

        for (key, val) in self.request.headers.get_all():
            headers[key] = val

        response: Optional[ModelResponse] = self.model.process_event(request, headers)

        if response is None:
            return

        runtime_metrics = response.metrics
        if runtime_metrics is not None:
            if validate_metrics(runtime_metrics):
                self.seldon_metrics.update(runtime_metrics, self.event_type)
            else:
                logging.error("Metrics returned are invalid: " + str(runtime_metrics))

        revent = create_cloud_event(
            response.data,
            self.event_type,
            self.event_source,
            event_id=event.EventID(),
            extensions=event.Extensions(),
        )

        if response.data is not None:
            # Create event from response if reply_url is active
            revent_headers, revent_data = http_marshaller.ToRequest(
                revent, converters.TypeBinary, json.dumps
            )

            if not self.reply_url == "":
                logging.debug(json.dumps(revent.Properties()))
                logging.debug("binary CloudEvent")
                for k, v in revent_headers.items():
                    logging.debug("{0}: {1}\r\n".format(k, v))
                logging.debug(revent_data)
                forward_request(revent_headers, revent_data, self.reply_url)

            self.set_header("Content-Type", "application/json")
            for headers in revent_headers:
                self.set_header(headers, revent_headers[headers])
            self.write(revent_data)


def create_cloud_event(
    data: dict,
    event_type: str,
    event_source: str,
    extensions: dict,
    event_id: str = None,
) -> v1.Event:
    """
    Create a CloudEvent

    Parameters
    ----------
    data
         The data to send
    event_type
         The CE event type
    event_source
         The CE event source
    extensions
         Any extensions to add
    event_id
         The event id
    Returns
    -------
         A CloudEvent

    """
    if event_id is None or event_id == "":
        event_id = uuid.uuid1().hex

    event = (
        v1.Event()
        .SetData(data)
        .SetEventID(event_id if event_id else str(uuid.uuid1().hex))
        .SetSource(event_source)
        .SetEventType(event_type)
        .SetExtensions(extensions)
    )
    return event

class LivenessHandler(tornado.web.RequestHandler):
    def get(self):
        self.write("Alive")


class ProtocolHandler(tornado.web.RequestHandler):
    def initialize(self, protocol: Protocol):
        self.protocol = protocol

    def get(self):
        self.write(str(self.protocol.value))


class MetricsHandler(tornado.web.RequestHandler):
    def initialize(self, seldon_metrics: SeldonMetrics):
        self.seldon_metrics = seldon_metrics

    def get(self):
        metrics, mimetype = self.seldon_metrics.generate_metrics()
        self.set_header("Content-Type", mimetype)
        self.write(metrics)
