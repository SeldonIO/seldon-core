import argparse
import os
import importlib
import json
import time
import logging
import multiprocessing as mp
import sys
import seldon_core.persistence as persistence
from distutils.util import strtobool
from seldon_core.flask_utils import ANNOTATIONS_FILE
import seldon_core.wrapper as seldon_microservice
from typing import Dict, Callable
from seldon_core.flask_utils import SeldonMicroserviceException

logger = logging.getLogger(__name__)

PARAMETERS_ENV_NAME = "PREDICTIVE_UNIT_PARAMETERS"
SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_SERVICE_PORT"
DEFAULT_PORT = 5000

DEBUG_PARAMETER = "SELDON_DEBUG"
DEBUG = False


def start_servers(target1: Callable, target2: Callable) -> None:
    """
    Start servers

    Parameters
    ----------
    target1
       Main flask process
    target2
       Auxilary flask process

    """
    p2 = mp.Process(target=target2)
    p2.daemon = True
    p2.start()

    target1()

    p2.join()


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
        "BOOL": bool
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
                    "Bad model parameter: " + name + " with value " + value + " can't be parsed as a " + type_,
                    reason="MICROSERVICE_BAD_PARAMETER")
            except KeyError:
                raise SeldonMicroserviceException(
                    "Bad model parameter type: " + type_ + " valid are INT, FLOAT, DOUBLE, STRING, BOOL",
                    reason="MICROSERVICE_BAD_PARAMETER")
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
                    parts = line.split("=")
                    if len(parts) == 2:
                        value = parts[1][1:-1]
                        logger.info("Found annotation %s:%s ", parts[0], value)
                        annotations[parts[0]] = value
                    else:
                        logger.info("bad annotation [%s]", line)
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
                'sampler': {
                    'type': 'const',
                    'param': 1,
                },
                'local_agent': {
                    'reporting_host': jaeger_serv,
                    'reporting_port': jaeger_port,
                },
                'logging': True,
            },
            service_name=interface_name,
            validate=True,
        )
    else:
        logger.info("Loading tracing config from %s", jaeger_config)
        import yaml
        with open(jaeger_config, 'r') as stream:
            config_dict = yaml.load(stream)
            config = Config(
                config=config_dict,
                service_name=interface_name,
                validate=True,
            )
    # this call also sets opentracing.tracer
    return config.initialize_tracer()


def main():
    LOG_FORMAT = '%(asctime)s - %(name)s:%(funcName)s:%(lineno)s - %(levelname)s:  %(message)s'
    logging.basicConfig(level=logging.INFO, format=LOG_FORMAT)
    logger.info('Starting microservice.py:main')

    sys.path.append(os.getcwd())
    parser = argparse.ArgumentParser()
    parser.add_argument("interface_name", type=str,
                        help="Name of the user interface.")
    parser.add_argument("api_type", type=str, choices=["REST", "GRPC", "FBS"])

    parser.add_argument("--service-type", type=str, choices=[
        "MODEL", "ROUTER", "TRANSFORMER", "COMBINER", "OUTLIER_DETECTOR"], default="MODEL")
    parser.add_argument("--persistence", nargs='?',
                        default=0, const=1, type=int)
    parser.add_argument("--parameters", type=str,
                        default=os.environ.get(PARAMETERS_ENV_NAME, "[]"))
    parser.add_argument("--log-level", type=str, default="INFO")
    parser.add_argument("--tracing", nargs='?',
                        default=int(os.environ.get("TRACING", "0")), const=1, type=int)

    args = parser.parse_args()

    parameters = parse_parameters(json.loads(args.parameters))

    # set up log level
    log_level_num = getattr(logging, args.log_level.upper(), None)
    if not isinstance(log_level_num, int):
        raise ValueError('Invalid log level: %s', args.log_level)

    logger.setLevel(log_level_num)
    logger.debug("Log level set to %s:%s", args.log_level, log_level_num)

    annotations = load_annotations()
    logger.info("Annotations: %s", annotations)

    parts = args.interface_name.rsplit(".", 1)
    if len(parts) == 1:
        logger.info("Importing %s",args.interface_name)
        interface_file = importlib.import_module(args.interface_name)
        user_class = getattr(interface_file, args.interface_name)
    else:
        logger.info("Importing submodule %s",parts)
        interface_file = importlib.import_module(parts[0])
        user_class = getattr(interface_file, parts[1])

    if args.persistence:
        logger.info('Restoring persisted component')
        user_object = persistence.restore(user_class, parameters)
        persistence.persist(user_object, parameters.get("push_frequency"))
    else:
        user_object = user_class(**parameters)

    # set log level for the imported microservice type
    seldon_microservice.logger.setLevel(log_level_num)
    logging.getLogger().setLevel(log_level_num)
    for handler in logger.handlers:
        handler.setLevel(log_level_num)

    port = int(os.environ.get(SERVICE_PORT_ENV_NAME, DEFAULT_PORT))

    if args.tracing:
        tracer = setup_tracing(args.interface_name)

    if args.api_type == "REST":

        def rest_prediction_server():
            app = seldon_microservice.get_rest_microservice(user_object)

            if args.tracing:
                from flask_opentracing import FlaskTracer
                tracing = FlaskTracer(tracer, True, app)

            app.run(host='0.0.0.0', port=port)

        logger.info("REST microservice running on port %i", port)
        server1_func = rest_prediction_server

    elif args.api_type == "GRPC":
        def grpc_prediction_server():

            if args.tracing:
                from grpc_opentracing import open_tracing_server_interceptor
                logger.info("Adding tracer")
                interceptor = open_tracing_server_interceptor(tracer)
            else:
                interceptor = None

            server = seldon_microservice.get_grpc_server(
                user_object, annotations=annotations, trace_interceptor=interceptor)

            server.add_insecure_port(f"0.0.0.0:{port}")

            server.start()

            logger.info("GRPC microservice Running on port %i", port)
            while True:
                time.sleep(1000)

        server1_func = grpc_prediction_server

    else:
        server1_func = None

    if hasattr(user_object, 'custom_service') and callable(getattr(user_object, 'custom_service')):
        server2_func = user_object.custom_service
    else:
        server2_func = None

        logger.info('Starting servers')
    start_servers(server1_func, server2_func)


if __name__ == "__main__":
    main()
