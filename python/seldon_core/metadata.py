import os
import json
import logging

from typing import Dict

from seldon_core.metrics import split_image_tag

from jsonschema import validate
from jsonschema.exceptions import ValidationError


logger = logging.getLogger(__name__)

MODEL_IMAGE = os.environ.get("PREDICTIVE_UNIT_IMAGE")


class SeldonInvalidMetadataError(Exception):
    pass


SELDON_ARRAY_SCHEMA = {
    "type": "object",
    "properties": {
        "datatype": {"type": "string", "enum": ["array"]},
        "shape": {"type": "array", "items": {"type": "integer"}},
    },
    "required": ["datatype", "shape"],
    "additionalProperties": False,
}

SELDON_JSON_SCHEMA = {
    "type": "object",
    "properties": {
        "datatype": {"type": "string", "enum": ["jsonData"]},
        "schema": {"type": "object"},
    },
    "required": ["datatype"],
    "additionalProperties": False,
}

SELDON_STR_SCHEMA = {
    "type": "object",
    "properties": {"datatype": {"type": "string", "enum": ["strData"]}},
    "required": ["datatype"],
    "additionalProperties": False,
}

SELDON_BIN_SCHEMA = {
    "type": "object",
    "properties": {"datatype": {"type": "string", "enum": ["binData"]}},
    "required": ["datatype"],
    "additionalProperties": False,
}

METADATA_TENSOR_SCHEMA = {
    "type": "array",
    "items": {
        "type": "object",
        "properties": {
            "datatype": {"type": "string"},
            "name": {"type": "string"},
            "shape": {"type": "array", "items": {"type": "integer"}},
        },
        "additionalProperties": False,
    },
    "additionalProperties": False,
}

V1_SCHEMA = {
    "type": "object",
    "properties": {
        "apiVersion": {"type": "string", "enum": ["v1"]},
        "name": {"type": "string"},
        "versions": {"type": "array", "items": {"type": "string"}},
        "platform": {"type": "string"},
        "inputs": {
            "oneOf": [
                SELDON_ARRAY_SCHEMA,
                SELDON_JSON_SCHEMA,
                SELDON_STR_SCHEMA,
                SELDON_BIN_SCHEMA,
            ]
        },
        "outputs": {
            "oneOf": [
                SELDON_ARRAY_SCHEMA,
                SELDON_JSON_SCHEMA,
                SELDON_STR_SCHEMA,
                SELDON_BIN_SCHEMA,
            ]
        },
    },
    "additionalProperties": False,
    "required": ["apiVersion"],
}

V2_SCHEMA = {
    "type": "object",
    "properties": {
        "apiVersion": {"type": "string", "enum": ["v2"]},
        "name": {"type": "string"},
        "versions": {"type": "array", "items": {"type": "string"}},
        "platform": {"type": "string"},
        "inputs": METADATA_TENSOR_SCHEMA,
        "outputs": METADATA_TENSOR_SCHEMA,
    },
    "additionalProperties": False,
}


def validate_model_metadata(data: Dict) -> Dict:
    """Validate metadata.

    Parameters
    ----------
    data
        User defined model metadata (json)

    Returns
    -------
        Validated model metadata (json)

    Raises
    ------
    SeldonInvalidMetadataError if data cannot be properly validated

    Notes
    -----

    Read data from json and validate against v1 or v2 metadata schema.
    SeldonInvalidMetadataError exception will be raised if validation fails.
    """
    if MODEL_IMAGE is not None:
        image_name, image_version = split_image_tag(MODEL_IMAGE)
    else:
        image_name, image_version = "", ""

    default_meta = {
        "apiVersion": "v2",
        "name": image_name,
        "versions": [image_version],
        "platform": "",
        "inputs": [],
        "outputs": [],
    }

    data = {**default_meta, **data}
    v = data.get("apiVersion", "v2")

    if v == "v1":
        schema = V1_SCHEMA
    elif v == "v2":
        schema = V2_SCHEMA
    else:
        raise SeldonInvalidMetadataError(f"Unknown metadata schema: {v}")

    try:
        validate(data, schema)
    except ValidationError as e:
        raise SeldonInvalidMetadataError(e)

    logger.debug(f"Successfully validated metadata:\n{json.dumps(data, indent=2)}")
    return data
