from seldon_core.metrics import validate_metrics
from seldon_core.microservice import SeldonMicroserviceException
import json
from seldon_core.proto import prediction_pb2
from typing import Dict, List, Union
import numpy as np

def client_custom_tags(user_model: object) -> Dict:
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
        return user_model.tags()
    else:
        return {}


def client_class_names(user_model: object, predictions: np.ndarray) -> List[str]:
    '''
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
    '''
    if len(predictions.shape) > 1:
        if hasattr(user_model, "class_names"):
            return user_model.class_names
        else:
            n_targets = predictions.shape[1]
            return ["t:{}".format(i) for i in range(n_targets)]
    else:
        return []


def client_predict(user_model: object, features: Union[np.ndarray,str,bytes], feature_names: List, **kwargs: Dict) -> Union[np.ndarray,List,str,bytes]:
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
    if hasattr(user_model,"predict"):
        try:
            return user_model.predict(features, feature_names, **kwargs)
        except TypeError:
            return user_model.predict(features, feature_names)
    else:
        return []


def client_transform_input(user_model: object, features: Union[np.ndarray,str,bytes], feature_names: List, **kwargs: Dict) -> Union[np.ndarray,List,str,bytes]:
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
            return user_model.transform_input(features, feature_names, **kwargs)
        except TypeError:
            return user_model.transform_input(features, feature_names)
    else:
        return features


def client_transform_output(user_model: object, features: Union[np.ndarray,str,bytes], feature_names: List, **kwargs: Dict) -> Union[np.ndarray,List,str,bytes]:
    """
    Transform output
    Parameters
    ----------
    user_model
       A Seldon user model
    features
       Data payload
    feature_names
       Data payload couln names
    kwargs
       Optional keyowrd args
    Returns
    -------
       Transformed data

    """
    if hasattr(user_model, "transform_output"):
        try:
            return user_model.transform_output(features, feature_names, **kwargs)
        except TypeError:
            return user_model.transform_output(features, feature_names)
    else:
        return features


def client_custom_metrics(user_model: object) -> List[Dict]:
    """
    Get custom metrics
    Parameters
    ----------
    user_model
       A Seldon user model

    Returns
    -------
       A list of custom metrics

    """
    if hasattr(user_model, "metrics"):
        metrics = user_model.metrics()
        if not validate_metrics(metrics):
            jStr = json.dumps(metrics)
            raise SeldonMicroserviceException(
                "Bad metric created during request: " + jStr, reason="MICROSERVICE_BAD_METRIC")
        return metrics
    else:
        return []


def client_feature_names(user_model: object, original: List) -> List:
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
        return user_model.feature_names
    else:
        return original


def client_send_feedback(user_model: object, features: Union[np.ndarray,str,bytes], feature_names: List, reward: float, truth: Union[np.ndarray,str,bytes], routing: Union[int,None]) -> Union[np.ndarray,List,str,bytes,None]:
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
        return user_model.send_feedback(features, feature_names, reward, truth, routing=routing)


def client_route(user_model: object, features: Union[np.ndarray,str,bytes], feature_names: List) -> int:
    """
    Gte routing from user model
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
        return user_model.route(features, feature_names)
    else:
        return -1


def client_aggregate(user_model: object, features_list: List[Union[np.ndarray,str,bytes]], feature_names_list: List) -> Union[np.ndarray,List,str,bytes]:
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
       An aggrehated payload
    """
    if hasattr(user_model, "aggregate"):
        return user_model.aggregate(features_list, feature_names_list)
    else:
        return features_list[0]