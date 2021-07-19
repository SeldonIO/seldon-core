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


def get_predictior_version() -> str:
    return json.loads(os.environ.get(ENV_PREDICTOR_LABELS, "{}")).get(
        "version", f"{NONIMPLEMENTED_MSG}")


def get_predictor_name() -> str:
    return os.environ.get(ENV_PREDICTOR_NAME, f"{NONIMPLEMENTED_MSG}")


def get_deployment_name() -> str:
    return os.environ.get(
        ENV_SELDON_DEPLOYMENT_NAME, f"{NONIMPLEMENTED_MSG}"
    )


def get_model_name() -> str:
    return os.environ.get(ENV_MODEL_NAME, NONIMPLEMENTED_MSG)


def get_image_name() -> str:
    return os.environ.get(
        ENV_MODEL_IMAGE, NONIMPLEMENTED_IMAGE_MSG)
