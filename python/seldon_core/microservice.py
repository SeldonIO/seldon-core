import argparse
import os
import importlib
import json
import time
import logging
import multiprocessing as mp
import threading
import sys

from typing import Dict, Callable
from distutils.util import strtobool

from seldon_core import persistence, __version__, wrapper as seldon_microservice
from seldon_core.metrics import SeldonMetrics
from seldon_core.flask_utils import ANNOTATIONS_FILE, SeldonMicroserviceException
from seldon_core.utils import getenv_as_bool, setup_tracing
from seldon_core.app import (
    StandaloneApplication,
    UserModelApplication,
    accesslog,
    threads,
    post_worker_init,
)

logger = logging.getLogger(__name__)

PARAMETERS_ENV_NAME = "PREDICTIVE_UNIT_PARAMETERS"
HTTP_SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_HTTP_SERVICE_PORT"
GRPC_SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_GRPC_SERVICE_PORT"
METRICS_SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_METRICS_SERVICE_PORT"

FILTER_METRICS_ACCESS_LOGS_ENV_NAME = "FILTER_METRICS_ACCESS_LOGS"

LOG_LEVEL_ENV = "SELDON_LOG_LEVEL"
DEFAULT_LOG_LEVEL = "INFO"

DEFAULT_GRPC_PORT = 5000
DEFAULT_HTTP_PORT = 9000
DEFAULT_METRICS_PORT = 6000

DEBUG_ENV = "SELDON_DEBUG"
GUNICORN_ACCESS_LOG_ENV = "GUNICORN_ACCESS_LOG"


def start_servers(
    target1: Callable, target2: Callable, target3: Callable, metrics_target: Callable
) -> None:
    """
    Start servers

    Parameters
    ----------
    target1
       Main flask process
    target2
       Auxilary flask process

    """
    p2 = None
    if target2:
        p2 = mp.Process(target=target2, daemon=True)
        p2.start()

    p3 = None
    if target3:
        p3 = mp.Process(target=target3, daemon=True)
        p3.start()

    p4 = None
    if metrics_target:
        p4 = mp.Process(target=metrics_target, daemon=True)
        p4.start()

    target1()

    if p2:
        p2.join()

    if p3:
        p3.join()

    if p4:
        p4.join()


def parse_parameters(parameters: Dict) -> Dict:
    """
    Parse the user object parameters

    Parameters
    ----------
    parameters

    Returns
    -------

    """
    type_dict = {
        "INT": int,
        "FLOAT": float,
        "DOUBLE": float,
        "STRING": str,
        "BOOL": bool,
    }
    parsed_parameters = {}
    for param in parameters:
        name = param.get("name")
        value = param.get("value")
        type_ = param.get("type")
        if type_ == "BOOL":
            parsed_parameters[name] = bool(strtobool(value))
        else:
            try:
                parsed_parameters[name] = type_dict[type_](value)
            except ValueError:
                raise SeldonMicroserviceException(
                    "Bad model parameter: "
                    + name
                    + " with value "
                    + value
                    + " can't be parsed as a "
                    + type_,
                    reason="MICROSERVICE_BAD_PARAMETER",
                )
            except KeyError:
                raise SeldonMicroserviceException(
                    "Bad model parameter type: "
                    + type_
                    + " valid are INT, FLOAT, DOUBLE, STRING, BOOL",
                    reason="MICROSERVICE_BAD_PARAMETER",
                )
    return parsed_parameters


def load_annotations() -> Dict:
    """
    Attempt to load annotations

    Returns
    -------

    """
    annotations = {}
    try:
        if os.path.isfile(ANNOTATIONS_FILE):
            with open(ANNOTATIONS_FILE, "r") as ins:
                for line in ins:
                    line = line.rstrip()
                    parts = list(map(str.strip, line.split("=", 1)))
                    if len(parts) == 2:
                        key = parts[0]
                        value = parts[1][1:-1]  # strip quotes at start and end
                        logger.info("Found annotation %s:%s ", key, value)
                        annotations[key] = value
                    else:
                        logger.info("Bad annotation [%s]", line)
    except:
        logger.error("Failed to open annotations file %s", ANNOTATIONS_FILE)
    return annotations


class MetricsEndpointFilter(logging.Filter):
    def filter(self, record):
        return seldon_microservice.METRICS_ENDPOINT not in record.getMessage()


def setup_logger(log_level: str, debug_mode: bool) -> logging.Logger:
    # set up log level
    log_level_raw = os.environ.get(LOG_LEVEL_ENV, log_level.upper())
    log_level_num = getattr(logging, log_level_raw, None)
    if not isinstance(log_level_num, int):
        raise ValueError("Invalid log level: %s", log_level)

    logger.setLevel(log_level_num)

    # Set right level on access logs
    flask_logger = logging.getLogger("werkzeug")
    flask_logger.setLevel(log_level_num)

    if getenv_as_bool(FILTER_METRICS_ACCESS_LOGS_ENV_NAME, default=not debug_mode):
        flask_logger.addFilter(MetricsEndpointFilter())
        gunicorn_logger = logging.getLogger("gunicorn.access")
        gunicorn_logger.addFilter(MetricsEndpointFilter())

    logger.debug("Log level set to %s:%s", log_level, log_level_num)

    # set log level for the imported microservice type
    seldon_microservice.logger.setLevel(log_level_num)
    logging.getLogger().setLevel(log_level_num)
    for handler in logger.handlers:
        handler.setLevel(log_level_num)

    return logger


def main():
    LOG_FORMAT = (
        "%(asctime)s - %(name)s:%(funcName)s:%(lineno)s - %(levelname)s:  %(message)s"
    )
    logging.basicConfig(level=logging.INFO, format=LOG_FORMAT)
    logger.info("Starting microservice.py:main")
    logger.info(f"Seldon Core version: {__version__}")

    sys.path.append(os.getcwd())
    parser = argparse.ArgumentParser()
    parser.add_argument("interface_name", type=str, help="Name of the user interface.")

    parser.add_argument(
        "--service-type",
        type=str,
        choices=["MODEL", "ROUTER", "TRANSFORMER", "COMBINER", "OUTLIER_DETECTOR"],
        default="MODEL",
    )
    parser.add_argument("--persistence", nargs="?", default=0, const=1, type=int)
    parser.add_argument(
        "--parameters", type=str, default=os.environ.get(PARAMETERS_ENV_NAME, "[]")
    )
    parser.add_argument(
        "--log-level",
        type=str,
        choices=["DEBUG", "INFO", "WARNING", "ERROR"],
        default=DEFAULT_LOG_LEVEL,
        help="Log level of the inference server.",
    )
    parser.add_argument(
        "--debug",
        nargs="?",
        type=bool,
        default=getenv_as_bool(DEBUG_ENV, default=False),
        const=True,
        help="Enable debug mode.",
    )
    parser.add_argument(
        "--tracing",
        nargs="?",
        default=int(os.environ.get("TRACING", "0")),
        const=1,
        type=int,
    )

    # gunicorn settings, defaults are from
    # http://docs.gunicorn.org/en/stable/settings.html
    parser.add_argument(
        "--workers",
        type=int,
        default=int(os.environ.get("GUNICORN_WORKERS", "1")),
        help="Number of Gunicorn workers for handling requests.",
    )
    parser.add_argument(
        "--threads",
        type=int,
        default=int(os.environ.get("GUNICORN_THREADS", "10")),
        help="Number of threads to run per Gunicorn worker.",
    )
    parser.add_argument(
        "--max-requests",
        type=int,
        default=int(os.environ.get("GUNICORN_MAX_REQUESTS", "0")),
        help="Maximum number of requests gunicorn worker will process before restarting.",
    )
    parser.add_argument(
        "--max-requests-jitter",
        type=int,
        default=int(os.environ.get("GUNICORN_MAX_REQUESTS_JITTER", "0")),
        help="Maximum random jitter to add to max-requests.",
    )

    parser.add_argument(
        "--single-threaded",
        type=int,
        default=int(os.environ.get("FLASK_SINGLE_THREADED", "0")),
        help="Force the Flask app to run single-threaded. Also applies to Gunicorn.",
    )

    parser.add_argument(
        "--http-port",
        type=int,
        default=int(os.environ.get(HTTP_SERVICE_PORT_ENV_NAME, DEFAULT_HTTP_PORT)),
        help="Set http port of seldon service",
    )

    parser.add_argument(
        "--grpc-port",
        type=int,
        default=int(os.environ.get(GRPC_SERVICE_PORT_ENV_NAME, DEFAULT_GRPC_PORT)),
        help="Set grpc port of seldon service",
    )

    parser.add_argument(
        "--metrics-port",
        type=int,
        default=int(
            os.environ.get(METRICS_SERVICE_PORT_ENV_NAME, DEFAULT_METRICS_PORT)
        ),
        help="Set metrics port of seldon service",
    )

    parser.add_argument(
        "--pidfile", type=str, default=None, help="A file path to use for the PID file"
    )

    parser.add_argument(
        "--access-log",
        nargs="?",
        type=bool,
        default=getenv_as_bool(GUNICORN_ACCESS_LOG_ENV, default=False),
        const=True,
        help="Enable gunicorn access log.",
    )

    args = parser.parse_args()
    parameters = parse_parameters(json.loads(args.parameters))

    setup_logger(args.log_level, args.debug)

    # set flask trace jaeger extra tags
    jaeger_extra_tags = list(
        filter(
            lambda x: (x != ""),
            [tag.strip() for tag in os.environ.get("JAEGER_EXTRA_TAGS", "").split(",")],
        )
    )
    logger.info("Parse JAEGER_EXTRA_TAGS %s", jaeger_extra_tags)

    annotations = load_annotations()
    logger.info("Annotations: %s", annotations)

    parts = args.interface_name.rsplit(".", 1)
    if len(parts) == 1:
        logger.info("Importing %s", args.interface_name)
        interface_file = importlib.import_module(args.interface_name)
        user_class = getattr(interface_file, args.interface_name)
    else:
        logger.info("Importing submodule %s", parts)
        interface_file = importlib.import_module(parts[0])
        user_class = getattr(interface_file, parts[1])

    if args.persistence:
        logger.info("Restoring persisted component")
        user_object = persistence.restore(user_class, parameters)
        persistence.persist(user_object, parameters.get("push_frequency"))
    else:
        user_object = user_class(**parameters)

    http_port = args.http_port
    grpc_port = args.grpc_port
    metrics_port = args.metrics_port

    # if args.tracing:
    #    tracer = setup_tracing(args.interface_name)

    seldon_metrics = SeldonMetrics(worker_id_func=os.getpid)
    # TODO why 2 ways to create metrics server
    # seldon_metrics = SeldonMetrics(
    #    worker_id_func=lambda: threading.current_thread().name
    # )
    if args.debug:
        # Start Flask debug server
        def rest_prediction_server():
            app = seldon_microservice.get_rest_microservice(user_object, seldon_metrics)
            try:
                user_object.load()
            except (NotImplementedError, AttributeError):
                pass
            if args.tracing:
                logger.info("Tracing branch is active")
                from flask_opentracing import FlaskTracing

                tracer = setup_tracing(args.interface_name)

                logger.info("Set JAEGER_EXTRA_TAGS %s", jaeger_extra_tags)
                FlaskTracing(tracer, True, app, jaeger_extra_tags)

            app.run(
                host="0.0.0.0",
                port=http_port,
                threaded=False if args.single_threaded else True,
            )

        logger.info(
            "REST microservice running on port %i single-threaded=%s",
            http_port,
            args.single_threaded,
        )
        server1_func = rest_prediction_server
    else:
        # Start production server
        def rest_prediction_server():
            options = {
                "bind": "%s:%s" % ("0.0.0.0", http_port),
                "accesslog": accesslog(args.access_log),
                "loglevel": args.log_level.lower(),
                "timeout": 5000,
                "threads": threads(args.threads, args.single_threaded),
                "workers": args.workers,
                "max_requests": args.max_requests,
                "max_requests_jitter": args.max_requests_jitter,
                "post_worker_init": post_worker_init,
            }
            if args.pidfile is not None:
                options["pidfile"] = args.pidfile
            app = seldon_microservice.get_rest_microservice(user_object, seldon_metrics)

            UserModelApplication(
                app,
                user_object,
                jaeger_extra_tags,
                args.interface_name,
                options=options,
            ).run()

        logger.info("REST gunicorn microservice running on port %i", http_port)
        server1_func = rest_prediction_server

    def grpc_prediction_server():

        if args.tracing:
            from grpc_opentracing import open_tracing_server_interceptor

            logger.info("Adding tracer")
            tracer = setup_tracing(args.interface_name)
            interceptor = open_tracing_server_interceptor(tracer)
        else:
            interceptor = None

        server = seldon_microservice.get_grpc_server(
            user_object,
            seldon_metrics,
            annotations=annotations,
            trace_interceptor=interceptor,
        )

        try:
            user_object.load()
        except (NotImplementedError, AttributeError):
            pass

        server.add_insecure_port(f"0.0.0.0:{grpc_port}")

        server.start()

        logger.info("GRPC microservice Running on port %i", grpc_port)
        while True:
            time.sleep(1000)

    server2_func = grpc_prediction_server

    def rest_metrics_server():
        app = seldon_microservice.get_metrics_microservice(seldon_metrics)
        if args.debug:
            app.run(host="0.0.0.0", port=metrics_port)
        else:
            options = {
                "bind": "%s:%s" % ("0.0.0.0", metrics_port),
                "accesslog": accesslog(args.access_log),
                "loglevel": args.log_level.lower(),
                "timeout": 5000,
                "max_requests": args.max_requests,
                "max_requests_jitter": args.max_requests_jitter,
                "post_worker_init": post_worker_init,
            }
            if args.pidfile is not None:
                options["pidfile"] = args.pidfile
            StandaloneApplication(app, options=options).run()

    logger.info("REST metrics microservice running on port %i", metrics_port)
    metrics_server_func = rest_metrics_server

    if hasattr(user_object, "custom_service") and callable(
        getattr(user_object, "custom_service")
    ):
        server3_func = user_object.custom_service
    else:
        server3_func = None

    logger.info("Starting servers")
    start_servers(server1_func, server2_func, server3_func, metrics_server_func)


if __name__ == "__main__":
    main()
