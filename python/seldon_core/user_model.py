from seldon_core.metrics import validate_metrics
from seldon_core.microservice import SeldonMicroserviceException
import json
from google.protobuf import json_format
from seldon_core.proto import prediction_pb2

def client_custom_tags(component):
    if hasattr(component, "tags"):
        return component.tags()
    else:
        return None

def client_class_names(user_model, predictions):
    '''

    Parameters
    ----------
    user_model
       User defined class instance
    predictions
       Prediction results
    Returns
    -------
       Class names
    '''
    n_targets = predictions.shape[1]
    if len(predictions.shape) > 1:
        if hasattr(user_model, "class_names"):
            return user_model.class_names
        else:
            return ["t:{}".format(i) for i in range(n_targets)]
    else:
        return []


def client_predict(user_model, features, feature_names, **kwargs):
    try:
        return user_model.predict(features, feature_names, **kwargs)
    except TypeError:
        return user_model.predict(features, feature_names)

def client_transform_input(user_model, features, feature_names, **kwargs):
    if hasattr(user_model, "transform_input"):
        try:
            return user_model.transform_input(features, feature_names, **kwargs)
        except TypeError:
            return user_model.transform_input(features, feature_names)
    else:
        return features


def client_transform_output(user_model, features, feature_names, **kwargs):
    if hasattr(user_model, "transform_output"):
        try:
            return user_model.transform_output(features, feature_names, **kwargs)
        except TypeError:
            return user_model.transform_output(features, feature_names)
    else:
        return features

def client_custom_metrics(component):
    if hasattr(component, "metrics"):
        metrics = component.metrics()
        if not validate_metrics(metrics):
            jStr = json.dumps(metrics)
            raise SeldonMicroserviceException(
                "Bad metric created during request: " + jStr, reason="MICROSERVICE_BAD_METRIC")
        return metrics
    else:
        return None


def client_feature_names(user_model: object, original: prediction_pb2.SeldonMessage) -> object:
    if hasattr(user_model, "feature_names"):
        return user_model.feature_names
    else:
        return original


def client_send_feedback(user_model, features, feature_names, reward, truth):
    if hasattr(user_model, "send_feedback"):
        user_model.send_feedback(features, feature_names, reward, truth)

