"""
Utilities to deal with Environment variables
"""
import json
import os

ENV_MODEL_NAME = "PREDICTIVE_UNIT_ID"
ENV_MODEL_IMAGE = "PREDICTIVE_UNIT_IMAGE"
ENV_SELDON_DEPLOYMENT_NAME = "SELDON_DEPLOYMENT_ID"
ENV_PREDICTOR_NAME = "PREDICTOR_ID"
ENV_PREDICTOR_LABELS = "PREDICTOR_LABELS"
NONIMPLEMENTED_MSG = "NOT_IMPLEMENTED"
NONIMPLEMENTED_IMAGE_MSG = f"{NONIMPLEMENTED_MSG}:{NONIMPLEMENTED_MSG}"


def get_predictor_version(default_val: str = NONIMPLEMENTED_MSG) -> str:
    """
    Get predictor version from `ENV_PREDICTOR_LABELS` environment variable.
    If not set return `default_val`


    Parameters
    ----------
    default_val
        Default value to return if the environment variable is not set

    Returns
    -------
       str

    """
    return json.loads(os.environ.get(ENV_PREDICTOR_LABELS, "{}")).get(
        "version", default_val
    )


def get_predictor_name(default_val: str = NONIMPLEMENTED_MSG) -> str:
    """
    Get predictor name from `ENV_PREDICTOR_NAME` environment variable.
    If not set return `default_val`


    Parameters
    ----------
    default_val
        Default value to return if the environment variable is not set

    Returns
    -------
       str

    """
    return os.environ.get(ENV_PREDICTOR_NAME, default_val)


def get_deployment_name(default_val: str = NONIMPLEMENTED_MSG) -> str:
    """
    Get deployment name from `ENV_SELDON_DEPLOYMENT_NAME` environment variable.
    If not set return `default_val`


    Parameters
    ----------
    default_val
        Default value to return if the environment variable is not set

    Returns
    -------
       str

    """
    return os.environ.get(ENV_SELDON_DEPLOYMENT_NAME, default_val)


def get_model_name(default_val: str = NONIMPLEMENTED_MSG) -> str:
    """
    Get model name from `ENV_MODEL_NAME` environment variable.
    If not set return `default_val`


    Parameters
    ----------
    default_val
        Default value to return if the environment variable is not set

    Returns
    -------
       str

    """
    return os.environ.get(ENV_MODEL_NAME, default_val)


def get_image_name(default_val: str = NONIMPLEMENTED_IMAGE_MSG) -> str:
    """
    Get model image name from `ENV_MODEL_IMAGE` environment variable.
    If not set return `default_val`


    Parameters
    ----------
    default_val
        Default value to return if the environment variable is not set

    Returns
    -------
       str

    """
    return os.environ.get(ENV_MODEL_IMAGE, default_val)
