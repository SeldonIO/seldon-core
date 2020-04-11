from seldon_core.metrics import validate_metrics
from seldon_core.flask_utils import SeldonMicroserviceException
import json
from typing import Dict, List, Union, Iterable
import numpy as np
from seldon_core.proto import prediction_pb2
from seldon_core.metrics import SeldonMetrics
import inspect
import logging
import os

logger = logging.getLogger(__name__)


INCLUDE_METRICS_IN_CLIENT_RESPONSE = (
    os.environ.get("INCLUDE_METRICS_IN_CLIENT_RESPONSE", "true").lower() == "true"
)


class SeldonNotImplementedError(SeldonMicroserviceException):
    status_code = 400

    def __init__(self, message):
        SeldonMicroserviceException.__init__(self, message)


class SeldonComponent(object):
    def __init__(self, **kwargs):
        pass

    def tags(self) -> Dict:
        raise SeldonNotImplementedError("tags is not implemented")

    def class_names(self) -> Iterable[str]:
        raise SeldonNotImplementedError("class_names is not implemented")

    def load(self):
        pass

    def predict(
        self, X: np.ndarray, names: Iterable[str], meta: Dict = None
    ) -> Union[np.ndarray, List, Dict, str, bytes]:
        raise SeldonNotImplementedError("predict is not implemented")

    def predict_raw(
        self, msg: prediction_pb2.SeldonMessage
    ) -> Union[prediction_pb2.SeldonMessage, Dict]:
        raise SeldonNotImplementedError("predict_raw is not implemented")

    def send_feedback_raw(
        self, feedback: prediction_pb2.Feedback
    ) -> Union[prediction_pb2.SeldonMessage, Dict]:
        raise SeldonNotImplementedError("send_feedback_raw is not implemented")

    def transform_input(
        self, X: np.ndarray, names: Iterable[str], meta: Dict = None
    ) -> Union[np.ndarray, List, Dict, str, bytes]:
        raise SeldonNotImplementedError("transform_input is not implemented")

    def transform_input_raw(
        self, msg: prediction_pb2.SeldonMessage
    ) -> Union[prediction_pb2.SeldonMessage, Dict]:
        raise SeldonNotImplementedError("transform_input_raw is not implemented")

    def transform_output(
        self, X: np.ndarray, names: Iterable[str], meta: Dict = None
    ) -> Union[np.ndarray, List, Dict, str, bytes]:
        raise SeldonNotImplementedError("transform_output is not implemented")

    def transform_output_raw(
        self, msg: prediction_pb2.SeldonMessage
    ) -> Union[prediction_pb2.SeldonMessage, Dict]:
        raise SeldonNotImplementedError("transform_output_raw is not implemented")

    def metrics(self) -> List[Dict]:
        raise SeldonNotImplementedError("metrics is not implemented")

    def feature_names(self) -> Iterable[str]:
        raise SeldonNotImplementedError("feature_names is not implemented")

    def send_feedback(
        self,
        features: Union[np.ndarray, str, bytes],
        feature_names: Iterable[str],
        reward: float,
        truth: Union[np.ndarray, str, bytes],
        routing: Union[int, None],
    ) -> Union[np.ndarray, List, Dict, str, bytes, None]:
        raise SeldonNotImplementedError("send_feedback is not implemented")

    def route(
        self, features: Union[np.ndarray, str, bytes], feature_names: Iterable[str]
    ) -> int:
        raise SeldonNotImplementedError("route is not implemented")

    def route_raw(
        self, msg: prediction_pb2.SeldonMessage
    ) -> Union[prediction_pb2.SeldonMessage, Dict]:
        raise SeldonNotImplementedError("route_raw is not implemented")

    def aggregate(
        self,
        features_list: List[Union[np.ndarray, str, bytes]],
        feature_names_list: List,
    ) -> Union[np.ndarray, List, Dict, str, bytes]:
        raise SeldonNotImplementedError("aggregate is not implemented")

    def aggregate_raw(
        self, msgs: prediction_pb2.SeldonMessageList
    ) -> Union[prediction_pb2.SeldonMessage, Dict]:
        raise SeldonNotImplementedError("aggregate_raw is not implemented")

    def health_status(self) -> Union[np.ndarray, List, str, bytes]:
        raise SeldonNotImplementedError("health is not implemented")

    def health_status_raw(self) -> prediction_pb2.SeldonMessage:
        raise SeldonNotImplementedError("health_raw is not implemented")

    def metadata(self) -> Dict:
        raise SeldonNotImplementedError("metadata is not implemented")


def client_custom_tags(user_model: SeldonComponent) -> Dict:
    """
    Get tags from user model

    Parameters
    ----------
    user_model

    Returns
    -------
       Dictionary of key value pairs

    """
    if hasattr(user_model, "tags"):
        try:
            return user_model.tags()
        except SeldonNotImplementedError:
            pass
    logger.debug("custom_tags is not implemented")
    return {}


def client_class_names(
    user_model: SeldonComponent, predictions: np.ndarray
) -> Iterable[str]:
    """
    Get class names from user model

    Parameters
    ----------
    user_model
       User defined class instance
    predictions
       Prediction results
    Returns
    -------
       Class names
    """
    if len(predictions.shape) > 1:
        if hasattr(user_model, "class_names"):
            if inspect.ismethod(getattr(user_model, "class_names")):
                try:
                    return user_model.class_names()
                except SeldonNotImplementedError:
                    pass
            else:
                logger.warning(
                    "class_names attribute is deprecated. Please define a class_names method"
                )
                return user_model.class_names
        logger.debug("class_names is not implemented")
        n_targets = predictions.shape[1]
        return ["t:{}".format(i) for i in range(n_targets)]
    else:
        return []


def client_predict(
    user_model: SeldonComponent,
    features: Union[np.ndarray, str, bytes],
    feature_names: Iterable[str],
    **kwargs: Dict
) -> Union[np.ndarray, List, str, bytes]:
    """
    Get prediction from user model

    Parameters
    ----------
    user_model
       A seldon user model
    features
       The data payload
    feature_names
       The feature names in the payload
    kwargs
       Optional keyword arguments
    Returns
    -------
       A prediction from the user model
    """
    if hasattr(user_model, "predict"):
        try:
            try:
                return user_model.predict(features, feature_names, **kwargs)
            except TypeError:
                return user_model.predict(features, feature_names)
        except SeldonNotImplementedError:
            pass
    logger.debug("predict is not implemented")
    return []


def client_transform_input(
    user_model: SeldonComponent,
    features: Union[np.ndarray, str, bytes],
    feature_names: Iterable[str],
    **kwargs: Dict
) -> Union[np.ndarray, List, str, bytes]:
    """
    Transform data with user model

    Parameters
    ----------
    user_model
       A Seldon user model
    features
       Data payload
    feature_names
       Data payload column names
    kwargs
       Optional keyword args

    Returns
    -------
       Transformed data

    """
    if hasattr(user_model, "transform_input"):
        try:
            try:
                return user_model.transform_input(features, feature_names, **kwargs)
            except TypeError:
                return user_model.transform_input(features, feature_names)
        except SeldonNotImplementedError:
            pass
    logger.debug("transform_input is not implemented")
    return features


def client_transform_output(
    user_model: SeldonComponent,
    features: Union[np.ndarray, str, bytes],
    feature_names: Iterable[str],
    **kwargs: Dict
) -> Union[np.ndarray, List, str, bytes]:
    """
    Transform output

    Parameters
    ----------
    user_model
       A Seldon user model
    features
       Data payload
    feature_names
       Data payload column names
    kwargs
       Optional keyword args
    Returns
    -------
       Transformed data

    """
    if hasattr(user_model, "transform_output"):
        try:
            try:
                return user_model.transform_output(features, feature_names, **kwargs)
            except TypeError:
                return user_model.transform_output(features, feature_names)
        except SeldonNotImplementedError:
            pass
    logger.debug("transform_output is not implemented")
    return features


def client_custom_metrics(
    user_model: SeldonComponent, seldon_metrics: SeldonMetrics
) -> List[Dict]:
    """
    Get custom metrics for client and update SeldonMetrics.

    This function will return empty list if INCLUDE_METRICS_IN_CLIENT_RESPONSE environmental
    variable is NOT set to "true" or "True".

    Parameters
    ----------
    user_model
       A Seldon user model
    seldon_metrics
        A SeldonMetrics instance

    Returns
    -------
       A list of custom metrics

    """
    if hasattr(user_model, "metrics"):
        try:
            metrics = user_model.metrics()
            if not validate_metrics(metrics):
                j_str = json.dumps(metrics)
                raise SeldonMicroserviceException(
                    "Bad metric created during request: " + j_str,
                    reason="MICROSERVICE_BAD_METRIC",
                )

            seldon_metrics.update(metrics)
            if INCLUDE_METRICS_IN_CLIENT_RESPONSE:
                return metrics
            else:
                return []
        except SeldonNotImplementedError:
            pass
    logger.debug("custom_metrics is not implemented")
    return []


def client_feature_names(
    user_model: SeldonComponent, original: Iterable[str]
) -> Iterable[str]:
    """
    Get feature names for user model

    Parameters
    ----------
    user_model
       A Seldon user model
    original
       Original feature names
    Returns
    -------
       A list if feature names
    """
    if hasattr(user_model, "feature_names"):
        try:
            return user_model.feature_names()
        except SeldonNotImplementedError:
            pass
    logger.debug("feature_names is not implemented")
    return original


def client_send_feedback(
    user_model: SeldonComponent,
    features: Union[np.ndarray, str, bytes],
    feature_names: Iterable[str],
    reward: float,
    truth: Union[np.ndarray, str, bytes],
    routing: Union[int, None],
) -> Union[np.ndarray, List, str, bytes, None]:
    """
    Feedback to user model

    Parameters
    ----------
    user_model
       A Seldon user model
    features
       A payload
    feature_names
       Payload column names
    reward
       Reward
    truth
       True outcome
    routing
       Optional routing

    Returns
    -------
       Optional payload

    """
    if hasattr(user_model, "send_feedback"):
        try:
            return user_model.send_feedback(
                features, feature_names, reward, truth, routing=routing
            )
        except SeldonNotImplementedError:
            pass
    logger.debug("send_feedback is not implemented")
    return None


def client_route(
    user_model: SeldonComponent,
    features: Union[np.ndarray, str, bytes],
    feature_names: Iterable[str],
    **kwargs: Dict
) -> int:
    """
    Get routing from user model

    Parameters
    ----------
    user_model
       A Seldon user model
    features
       Payload
    feature_names
       Columns for payload

    Returns
    -------
       Routing index for one of children
    """
    if hasattr(user_model, "route"):
        try:
            return user_model.route(features, feature_names, **kwargs)
        except TypeError:
            return user_model.route(features, feature_names)
    else:
        raise SeldonNotImplementedError("Route not defined")


def client_aggregate(
    user_model: SeldonComponent,
    features_list: List[Union[np.ndarray, str, bytes]],
    feature_names_list: List,
) -> Union[np.ndarray, List, str, bytes]:
    """
    Aggregate payloads

    Parameters
    ----------
    user_model
       A Seldon user model
    features_list
       A list of payloads
    feature_names_list
       Column names for payloads
    Returns
    -------
       An aggregated payload
    """
    if hasattr(user_model, "aggregate"):
        return user_model.aggregate(features_list, feature_names_list)
    else:
        raise SeldonNotImplementedError("Aggregate not defined")


def client_health_status(
    user_model: SeldonComponent,
) -> Union[np.ndarray, List, str, bytes]:
    """
    Perform a health check

    Parameters
    ----------
    user_model
       A Seldon user model
    Returns
    -------
       Health check results
    """
    if hasattr(user_model, "health_status"):
        return user_model.health_status()
    else:
        raise SeldonNotImplementedError("health_status not defined")
