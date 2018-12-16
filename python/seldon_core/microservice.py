from flask import Flask, Blueprint, request
import argparse
import numpy as np
import os
import importlib
import json
import time
import logging
import multiprocessing as mp
import tensorflow as tf
from tensorflow.core.framework.tensor_pb2 import TensorProto
from google.protobuf import json_format
from google.protobuf.struct_pb2 import ListValue
import sys

from seldon_core.proto import prediction_pb2
import seldon_core.persistence as persistence

logger = logging.getLogger(__name__)

PARAMETERS_ENV_NAME = "PREDICTIVE_UNIT_PARAMETERS"
SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_SERVICE_PORT"
DEFAULT_PORT = 5000

DEBUG_PARAMETER = "SELDON_DEBUG"
DEBUG = False

ANNOTATIONS_FILE = "/etc/podinfo/annotations"
ANNOTATION_GRPC_MAX_MSG_SIZE = 'seldon.io/grpc-max-message-size'            

def startServers(target1, target2):
    p2 = mp.Process(target=target2)
    p2.deamon = True
    p2.start()

    target1()

    p2.join()


class SeldonMicroserviceException(Exception):
    status_code = 400

    def __init__(self, message, status_code=None, payload=None, reason="MICROSERVICE_BAD_DATA"):
        Exception.__init__(self)
        self.message = message
        if status_code is not None:
            self.status_code = status_code
        self.payload = payload
        self.reason = reason

    def to_dict(self):
        rv = {"status": {"status": 1, "info": self.message,
                         "code": -1, "reason": self.reason}}
        return rv


def sanity_check_request(req):
    if not type(req) == dict:
        raise SeldonMicroserviceException("Request must be a dictionary")
    if "data" in req:
        data = req.get("data")
        if not type(data) == dict:
            raise SeldonMicroserviceException(
                "data field must be a dictionary")
        if data.get('ndarray') is None and data.get('tensor') is None and data.get('tftensor') is None:
            raise SeldonMicroserviceException(
                "Data dictionary has no 'tensor', 'ndarray' or 'tftensor' keyword.")
    elif not ("binData" in req or "strData" in req):
        raise SeldonMicroserviceException("Request must contain Default Data or binData or strData")
    # TODO: Should we check more things? Like shape not being None or empty for a tensor?


def extract_message():
    jStr = request.form.get("json")
    if jStr:
        message = json.loads(jStr)
    else:
        jStr = request.args.get('json')
        if jStr:
            message = json.loads(jStr)
        else:
            raise SeldonMicroserviceException("Empty json parameter in data")
    if message is None:
        raise SeldonMicroserviceException("Invalid Data Format")
    return message


def get_custom_tags(component):
    if hasattr(component, "tags"):
        return component.tags()
    else:
        return None


def array_to_list_value(array, lv=None):
    if lv is None:
        lv = ListValue()
    if len(array.shape) == 1:
        lv.extend(array)
    else:
        for sub_array in array:
            sub_lv = lv.add_list()
            array_to_list_value(sub_array, sub_lv)
    return lv


def get_data_from_json(message):
    if "data" in message:
        datadef = message.get("data")
        return rest_datadef_to_array(datadef)
    elif "binData" in message:
        return message["binData"]
    elif "strData" in message:
        return message["strData"]
    else:
        strJson = json.dumps(message)
        raise SeldonMicroserviceException(
            "Can't find data in json: " + strJson)


def rest_datadef_to_array(datadef):
    if datadef.get("tensor") is not None:
        features = np.array(datadef.get("tensor").get("values")).reshape(
            datadef.get("tensor").get("shape"))
    elif datadef.get("ndarray") is not None:
        features = np.array(datadef.get("ndarray"))
    elif datadef.get("tftensor") is not None:
        tfp = TensorProto()
        json_format.ParseDict(datadef.get("tftensor"),
                              tfp, ignore_unknown_fields=False)
        features = tf.make_ndarray(tfp)
    else:
        features = np.array([])
    return features


def array_to_rest_datadef(array, names, original_datadef):
    datadef = {"names": names}
    if original_datadef.get("tensor") is not None:
        datadef["tensor"] = {
            "shape": array.shape,
            "values": array.ravel().tolist()
        }
    elif original_datadef.get("ndarray") is not None:
        datadef["ndarray"] = array.tolist()
    elif original_datadef.get("tftensor") is not None:
        tftensor = tf.make_tensor_proto(array)
        jStrTensor = json_format.MessageToJson(tftensor)
        jTensor = json.loads(jStrTensor)
        datadef["tftensor"] = jTensor
    else:
        datadef["ndarray"] = array.tolist()
    return datadef


def get_data_from_proto(request):
    data_type = request.WhichOneof("data_oneof")
    if data_type == "data":
        datadef = request.data
        return grpc_datadef_to_array(datadef)
    elif data_type == "binData":
        return request.binData
    elif data_type == "strData":
        return request.strData
    else:
        raise SeldonMicroserviceException("Unknown data in SeldonMessage")


def grpc_datadef_to_array(datadef):
    data_type = datadef.WhichOneof("data_oneof")
    if data_type == "tensor":
        if (sys.version_info >= (3, 0)):
            sz = np.prod(datadef.tensor.shape)  # get number of float64 entries
            c = datadef.tensor.SerializeToString()  # get bytes
            # create array from packed entries which are at end of bytes - assumes same endianness
            features = np.frombuffer(memoryview(
                c[-(sz * 8):]), dtype=np.float64, count=sz, offset=0)
            features = features.reshape(datadef.tensor.shape)
        else:
            # Python 2 version which is slower
            features = np.array(datadef.tensor.values).reshape(
                datadef.tensor.shape)
    elif data_type == "ndarray":
        features = np.array(datadef.ndarray)
    elif data_type == "tftensor":
        features = tf.make_ndarray(datadef.tftensor)
    else:
        features = np.array([])
    return features


def array_to_grpc_datadef(array, names, data_type):
    if data_type == "tensor":
        datadef = prediction_pb2.DefaultData(
            names=names,
            tensor=prediction_pb2.Tensor(
                shape=array.shape,
                values=array.ravel().tolist()
            )
        )
    elif data_type == "ndarray":
        datadef = prediction_pb2.DefaultData(
            names=names,
            ndarray=array_to_list_value(array)
        )
    elif data_type == "tftensor":
        datadef = prediction_pb2.DefaultData(
            names=names,
            tftensor=tf.make_tensor_proto(array)
        )
    else:
        datadef = prediction_pb2.DefaultData(
            names=names,
            ndarray=array_to_list_value(array)
        )

    return datadef


def parse_parameters(parameters):
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
        parsed_parameters[name] = type_dict[type_](value)
    return parsed_parameters


def load_annotations():
    annotations = {}
    try:
        if os.path.isfile(ANNOTATIONS_FILE):
            with open(ANNOTATIONS_FILE, "r") as ins:
                for line in ins:
                    line = line.rstrip()
                    parts = line.split("=")
                    if len(parts) == 2:
                        value = parts[1]
                        value = parts[1][1:-1]
                        logger.info("Found annotation %s:%s ", parts[0], value)
                        annotations[parts[0]] = value
                    else:
                        logger.info("bad annotation [%s]", line)
    except:
        logger.error("Failed to open annotations file %s", ANNOTATIONS_FILE)
    return annotations


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
    parser.add_argument("--log-level", type=str, default='INFO')
    args = parser.parse_args()

    parameters = parse_parameters(json.loads(args.parameters))

    # set up log level
    log_level_num = getattr(logging, args.log_level.upper(), None)
    if not isinstance(log_level_num, int):
        raise ValueError('Invalid log level: %s', args.log_level)
    logger.setLevel(log_level_num)

    DEBUG = False
    if parameters.get(DEBUG_PARAMETER):
        parameters.pop(DEBUG_PARAMETER)
        DEBUG = True

    annotations = load_annotations()
    logger.info("Annotations: %s", annotations)

    interface_file = importlib.import_module(args.interface_name)
    user_class = getattr(interface_file, args.interface_name)

    if args.persistence:
        logger.info('Restoring persisted component')
        user_object = persistence.restore(user_class, parameters, debug=DEBUG)
        persistence.persist(user_object, parameters.get("push_frequency"))
    else:
        user_object = user_class(**parameters)

    if args.service_type == "MODEL":
        import seldon_core.model_microservice as seldon_microservice
    elif args.service_type == "ROUTER":
        import seldon_core.router_microservice as seldon_microservice
    elif args.service_type == "TRANSFORMER":
        import seldon_core.transformer_microservice as seldon_microservice
    elif args.service_type == "COMBINER":
        import seldon_core.combiner_microservice as seldon_microservice
    elif args.service_type == "OUTLIER_DETECTOR":
        import seldon_core.outlier_detector_microservice as seldon_microservice

    port = int(os.environ.get(SERVICE_PORT_ENV_NAME, DEFAULT_PORT))

    if args.api_type == "REST":
        def rest_prediction_server():
            app = seldon_microservice.get_rest_microservice(
                user_object, debug=DEBUG)
            app.run(host='0.0.0.0', port=port)

        logger.info("REST microservice running on port %i",port)
        server1_func = rest_prediction_server

    elif args.api_type == "GRPC":
        def grpc_prediction_server():
            server = seldon_microservice.get_grpc_server(
                user_object, debug=DEBUG, annotations=annotations)
            server.add_insecure_port("0.0.0.0:{}".format(port))
            server.start()

            logger.info("GRPC microservice Running on port %i",port)
            while True:
                time.sleep(1000)

        server1_func = grpc_prediction_server

    elif args.api_type == "FBS":
        def fbs_prediction_server():
            seldon_microservice.run_flatbuffers_server(user_object, port)

        logger.info("FBS microservice Running on port %i",port)
        server1_func = fbs_prediction_server

    else:
        server1_func = None

    if hasattr(user_object, 'custom_service') and callable(getattr(user_object, 'custom_service')):
        server2_func = user_object.custom_service
    else:
        server2_func = None

        logger.info('Starting servers')
    startServers(server1_func, server2_func)


if __name__ == "__main__":
    main()
