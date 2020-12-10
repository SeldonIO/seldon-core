import argparse

import tensorflow as tf
from enum import Enum
import os

tf.keras.backend.clear_session()
import logging
from adserver.cm_model import CustomMetricsModel
from adserver.od_model import AlibiDetectOutlierModel
from adserver.ad_model import AlibiDetectAdversarialDetectionModel
from adserver.cd_model import AlibiDetectConceptDriftModel
from adserver.server import CEServer
from adserver.protocols import Protocol
from adserver.server import DEFAULT_HTTP_PORT
from alibi_detect.utils.saving import Data


class AlibiDetectMethod(Enum):
    adversarial_detector = "AdversarialDetector"
    outlier_detector = "OutlierDetector"
    drift_detector = "DriftDetector"
    metrics_server = "MetricsServer"

    def __str__(self):
        return self.value


class GroupedAction(argparse.Action):  # pylint:disable=too-few-public-methods
    def __call__(self, theparser, namespace, values, option_string=None):
        group, dest = self.dest.split(".", 2)
        groupspace = getattr(namespace, group, argparse.Namespace())
        setattr(groupspace, dest, values)
        setattr(namespace, group, groupspace)


def str2bool(v):
    if isinstance(v, bool):
        return v
    if v.lower() in ("yes", "true", "t", "y", "1"):
        return True
    elif v.lower() in ("no", "false", "f", "n", "0"):
        return False
    else:
        raise argparse.ArgumentTypeError("Boolean value expected.")


DEFAULT_MODEL_NAME = "model"

parser = argparse.ArgumentParser(add_help=False)
parser.add_argument(
    "--http_port",
    default=DEFAULT_HTTP_PORT,
    type=int,
    help="The HTTP Port listened to by the model server.",
)
parser.add_argument(
    "--protocol",
    type=Protocol,
    choices=list(Protocol),
    default="tensorflow.http",
    help="The protocol served by the model server",
)
parser.add_argument(
    "--reply_url", type=str, default="", help="URL to send reply cloudevent"
)
parser.add_argument(
    "--event_source", type=str, default="", help="URI of the event source"
)
parser.add_argument(
    "--event_type",
    type=str,
    default="",
    help="e.g. io.seldon.serving.inference.outlier or org.kubeflow.serving.inference.outlier",
)
parser.add_argument(
    "--model_name",
    default=DEFAULT_MODEL_NAME,
    help="The name that the model is served under.",
)
parser.add_argument("--storage_uri", required=True, help="A URI pointer to the model")
parser.add_argument(
    "--elasticsearch_uri",
    type=str,
    help="A URI pointer to the elasticsearch database if relevant",
)

subparsers = parser.add_subparsers(help="sub-command help", dest="command")

# Concept Drift Arguments
parser_drift = subparsers.add_parser(str(AlibiDetectMethod.drift_detector))
parser_drift.add_argument(
    "--drift_batch_size",
    type=int,
    action=GroupedAction,
    dest="alibi.drift_batch_size",
    default=argparse.SUPPRESS,
)

parser_adversarial = subparsers.add_parser(str(AlibiDetectMethod.adversarial_detector))
parser_outlier = subparsers.add_parser(str(AlibiDetectMethod.outlier_detector))
parser_metrics = subparsers.add_parser(str(AlibiDetectMethod.metrics_server))

args, _ = parser.parse_known_args()

argdDict = vars(args).copy()
if "alibi" in argdDict:
    extra = vars(args.alibi)
else:
    extra = {}
logging.info("Extra args: %s", extra)

if __name__ == "__main__":
    method = AlibiDetectMethod(args.command)
    model: Data = None
    if method == AlibiDetectMethod.outlier_detector:
        model = AlibiDetectOutlierModel(args.model_name, args.storage_uri)
    elif method == AlibiDetectMethod.adversarial_detector:
        model = AlibiDetectAdversarialDetectionModel(args.model_name, args.storage_uri)
    elif method == AlibiDetectMethod.drift_detector:
        model = AlibiDetectConceptDriftModel(args.model_name, args.storage_uri, **extra)
    elif method == AlibiDetectMethod.metrics_server:
        model = CustomMetricsModel(
            args.model_name,
            args.storage_uri,
            elasticsearch_uri=args.elasticsearch_uri,
            **extra
        )
    else:
        logging.error("Unknown method %s", args.command)
        os._exit(-1)
    CEServer(
        args.protocol,
        args.event_type,
        args.event_source,
        http_port=args.http_port,
        reply_url=args.reply_url,
    ).start(model)
