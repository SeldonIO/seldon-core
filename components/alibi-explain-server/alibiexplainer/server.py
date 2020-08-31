import json
import logging
from http import HTTPStatus
from typing import Dict, Optional
import tornado.httpserver
import tornado.ioloop
import tornado.web
from alibiexplainer.model import ExplainerModel
from alibiexplainer.constants import SELDON_LOGLEVEL
import argparse

logging.basicConfig(level=SELDON_LOGLEVEL)

DEFAULT_HTTP_PORT = 8080
DEFAULT_GRPC_PORT = 8081
DEFAULT_MAX_BUFFER_SIZE = 104857600

server_parser = argparse.ArgumentParser(add_help=False)
server_parser.add_argument('--http_port', default=DEFAULT_HTTP_PORT, type=int,
                           help='The HTTP Port listened to by the model server.')
server_parser.add_argument('--grpc_port', default=DEFAULT_GRPC_PORT, type=int,
                           help='The GRPC Port listened to by the model server.')
server_parser.add_argument('--max_buffer_size', default=DEFAULT_MAX_BUFFER_SIZE, type=int,
                           help='The max buffer size for tornado.')
server_parser.add_argument('--workers', default=0, type=int,
                           help='The number of works to fork')
args, _ = server_parser.parse_known_args()

class ExplainerServer(object):
    def __init__(
        self,
        http_port: int = DEFAULT_HTTP_PORT
    ):
        """
        Explainer Server

        Parameters
        ----------
        http_port
             http port to listen on
        """
        self.registered_model: Optional[ExplainerModel] = None
        self.http_port = http_port
        self._http_server: Optional[tornado.httpserver.HTTPServer] = None

    def create_application(self):
        return tornado.web.Application(
            [
                (r"/v1/models/[a-zA-Z0-9_-]+:explain",
                 ExplainHandler, dict(model=self.registered_model)),
                (r"/api/v0.1/explain",
                 ExplainHandler, dict(model=self.registered_model)),
                (r"/api/v1.0/explain",
                 ExplainHandler, dict(model=self.registered_model)),
            ]
        )

    def start(self, model: ExplainerModel):
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

    def register_model(self, model: ExplainerModel):
        if not model.name:
            raise Exception("Failed to register model, model.name must be provided.")
        self.registered_model = model
        logging.info("Registering model:" + model.name)

class ExplainHandler(tornado.web.RequestHandler):
    def initialize(self, model: ExplainerModel):
        self.model = model  # pylint:disable=attribute-defined-outside-init

    def post(self):
        try:
            body = json.loads(self.request.body)
        except json.decoder.JSONDecodeError as e:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason="Unrecognized request format: %s" % e
            )
        response =self.model.explain(body)
        self.write(response)


