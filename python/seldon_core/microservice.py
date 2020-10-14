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
from seldon_core.utils import getenv_as_bool
from seldon_core.app import (
    StandaloneApplication,
    UserModelApplication,
    accesslog,
    threads,
    post_worker_init,
)

logger = logging.getLogger(__name__)

PARAMETERS_ENV_NAME = "PREDICTIVE_UNIT_PARAMETERS"
SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_SERVICE_PORT"
METRICS_SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_METRICS_SERVICE_PORT"

FILTER_METRICS_ACCESS_LOGS_ENV_NAME = "FILTER_METRICS_ACCESS_LOGS"

LOG_LEVEL_ENV = "SELDON_LOG_LEVEL"
DEFAULT_LOG_LEVEL = "INFO"

DEFAULT_PORT = 5000
DEFAULT_METRICS_PORT = 6000

DEBUG_ENV = "SELDON_DEBUG"


def start_servers(
    target1: Callable, target2: Callable, metrics_target: Callable
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
    if metrics_target:
        p3 = mp.Process(target=metrics_target, daemon=True)
        p3.start()

    target1()

    if p2:
        p2.join()

    if p3:
        p3.join()


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


def setup_tracing(interface_name: str) -> object:
    logger.info("Initializing tracing")
    from jaeger_client import Config

    jaeger_serv = os.environ.get("JAEGER_AGENT_HOST", "0.0.0.0")
    jaeger_port = os.environ.get("JAEGER_AGENT_PORT", 5775)
    jaeger_config = os.environ.get("JAEGER_CONFIG_PATH", None)
    if jaeger_config is None:
        logger.info("Using default tracing config")
        config = Config(
            config={  # usually read from some yaml config
                "sampler": {"type": "const", "param": 1},
                "local_agent": {
                    "reporting_host": jaeger_serv,
                    "reporting_port": jaeger_port,
                },
                "logging": True,
            },
            service_name=interface_name,
            validate=True,
        )
    else:
        logger.info("Loading tracing config from %s", jaeger_config)
        import yaml

        with open(jaeger_config, "r") as stream:
            config_dict = yaml.load(stream)
            config = Config(
                config=config_dict, service_name=interface_name, validate=True
            )
    # this call also sets opentracing.tracer
    return config.initialize_tracer()


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
    parser.add_argument("api_type", type=str, choices=["REST", "GRPC", "FBS"])

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
        "--port",
        type=int,
        default=int(os.environ.get(SERVICE_PORT_ENV_NAME, DEFAULT_PORT)),
        help="Set port of seldon service",
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
        "--pidfile",
        type=str,
        default=None,
        help="A filepath to use for the PID file",
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

    port = args.port
    metrics_port = args.metrics_port

    if args.tracing:
        tracer = setup_tracing(args.interface_name)

    if args.api_type == "REST":
        seldon_metrics = SeldonMetrics(worker_id_func=os.getpid)

        if args.debug:
            # Start Flask debug server
            def rest_prediction_server():
                app = seldon_microservice.get_rest_microservice(
                    user_object, seldon_metrics
                )
                try:
                    user_object.load()
                except (NotImplementedError, AttributeError):
                    pass
                if args.tracing:
                    logger.info("Tracing branch is active")
                    from flask_opentracing import FlaskTracing

                    logger.info("Set JAEGER_EXTRA_TAGS %s", jaeger_extra_tags)
                    FlaskTracing(tracer, True, app, jaeger_extra_tags)

                app.run(
                    host="0.0.0.0",
                    port=port,
                    threaded=False if args.single_threaded else True,
                )

            logger.info(
                "REST microservice running on port %i single-threaded=%s",
                port,
                args.single_threaded,
            )
            server1_func = rest_prediction_server
        else:
            # Start production server
            def rest_prediction_server():
                options = {
                    "bind": "%s:%s" % ("0.0.0.0", port),
                    "accesslog": accesslog(args.log_level),
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
                app = seldon_microservice.get_rest_microservice(
                    user_object, seldon_metrics
                )
                UserModelApplication(app, user_object, options=options).run()

            logger.info("REST gunicorn microservice running on port %i", port)
            server1_func = rest_prediction_server

    elif args.api_type == "GRPC":
        seldon_metrics = SeldonMetrics(
            worker_id_func=lambda: threading.current_thread().name
        )

        def grpc_prediction_server():

            if args.tracing:
                from grpc_opentracing import open_tracing_server_interceptor

                logger.info("Adding tracer")
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

            server.add_insecure_port(f"0.0.0.0:{port}")

            server.start()

            logger.info("GRPC microservice Running on port %i", port)
            while True:
                time.sleep(1000)

        server1_func = grpc_prediction_server

    else:
        server1_func = None

    def rest_metrics_server():
        app = seldon_microservice.get_metrics_microservice(seldon_metrics)
        if args.debug:
            app.run(host="0.0.0.0", port=metrics_port)
        else:
            options = {
                "bind": "%s:%s" % ("0.0.0.0", metrics_port),
                "accesslog": accesslog(args.log_level),
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
        server2_func = user_object.custom_service
    else:
        server2_func = None

    logger.info("Starting servers")
    start_servers(server1_func, server2_func, metrics_server_func)


if __name__ == "__main__":
    main()
