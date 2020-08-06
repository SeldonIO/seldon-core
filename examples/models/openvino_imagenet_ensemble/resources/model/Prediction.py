import numpy as np
import logging
import datetime
import os
import sys
import boto3
from urllib.parse import urlparse
from google.cloud import storage
from openvino.inference_engine import IENetwork, IEPlugin


def get_logger(name):
    logger = logging.getLogger(name)
    log_formatter = logging.Formatter("%(asctime)s - %(name)s - "
                                      "%(levelname)s - %(message)s")
    logger.setLevel('DEBUG')

    console_handler = logging.StreamHandler()
    console_handler.setFormatter(log_formatter)
    logger.addHandler(console_handler)

    return logger

logger = get_logger(__name__)


def gs_download_file(path):
    if path is None:
        return None
    parsed_path = urlparse(path)
    bucket_name = parsed_path.netloc
    file_path = parsed_path.path[1:]
    try:
        gs_client = storage.Client()
        bucket = gs_client.get_bucket(bucket_name)
    except:
        gs_client = storage.Client.create_anonymous_client()
        bucket = gs_client.bucket(bucket_name, user_project=None)
    blob = bucket.blob(file_path)
    tmp_path = os.path.join('/tmp', file_path.split(os.sep)[-1])
    blob.download_to_filename(tmp_path)
    return tmp_path


def s3_download_file(path):
    if path is None:
        return None
    s3_endpoint = os.getenv('S3_ENDPOINT')
    s3_client = boto3.client('s3', endpoint_url=s3_endpoint)
    parsed_path = urlparse(path)
    bucket_name = parsed_path.netloc
    file_path = parsed_path.path[1:]
    tmp_path = os.path.join('/tmp', file_path.split(os.sep)[-1])
    s3_transfer = boto3.s3.transfer.S3Transfer(s3_client)
    s3_transfer.download_file(bucket_name, file_path, tmp_path)
    return tmp_path


def GetLocalPath(requested_path):
    print("Trying to download ",requested_path)
    parsed_path = urlparse(requested_path)
    if parsed_path.scheme == '':
        return requested_path
    elif parsed_path.scheme == 'gs':
        return gs_download_file(path=requested_path)
    elif parsed_path.scheme == 's3':
        return s3_download_file(path=requested_path)


class Prediction(object):
    def __init__(self):
        try:
            xml_path = os.environ["XML_PATH"]
            bin_path = os.environ["BIN_PATH"]

        except KeyError:
            print("Please set the environment variables XML_PATH, BIN_PATH")
            sys.exit(1)

        xml_local_path = GetLocalPath(xml_path)
        bin_local_path = GetLocalPath(bin_path)
        print('path object', xml_local_path)

        CPU_EXTENSION = os.getenv('CPU_EXTENSION', "/usr/local/lib/libcpu_extension.so")

        plugin = IEPlugin(device='CPU', plugin_dirs=None)
        if CPU_EXTENSION:
            plugin.add_cpu_extension(CPU_EXTENSION)
        net = IENetwork(model=xml_local_path, weights=bin_local_path)
        self.input_blob = next(iter(net.inputs))
        self.out_blob = next(iter(net.outputs))
        self.batch_size = net.inputs[self.input_blob].shape[0]
        self.inputs = net.inputs
        self.outputs = net.outputs
        self.exec_net = plugin.load(network=net, num_requests=self.batch_size)


    def predict(self,X,feature_names):
        start_time = datetime.datetime.now()
        results = self.exec_net.infer(inputs={self.input_blob: X})
        predictions = results[self.out_blob]
        end_time = datetime.datetime.now()
        duration = (end_time - start_time).total_seconds() * 1000
        logger.debug("Processing time: {:.2f} ms".format(duration))
        return predictions.astype(np.float64)

