import json
import logging
import os
from typing import Dict

from jsonschema import validate
from jsonschema.exceptions import ValidationError

from seldon_core.metrics import split_image_tag

logger = logging.getLogger(__name__)

MODEL_IMAGE = os.environ.get("PREDICTIVE_UNIT_IMAGE")


class SeldonInvalidMetadataError(Exception):
    pass


SELDON_ARRAY_SCHEMA = {
    "type": "object",
    "properties": {
        "messagetype": {"type": "string", "enum": ["tensor", "ndarray", "tftensor"]},
        "schema": {
            "type": "object",
            "properties": {
                "names": {"type": "array", "items": {"type": "string"}},
                "shape": {"type": "array", "items": {"type": "integer"}},
                # Shall we also include field for datatype (as in dtype for np.array)?
            },
            "additionalProperties": False,
        },
    },
    "required": ["messagetype"],
    "additionalProperties": False,
}

SELDON_JSON_SCHEMA = {
    "type": "object",
    "properties": {
        "messagetype": {"type": "string", "enum": ["jsonData"]},
        "schema": {"type": "object"},
    },
    "required": ["messagetype"],
    "additionalProperties": False,
}

SELDON_STR_SCHEMA = {
    "type": "object",
    "properties": {"messagetype": {"type": "string", "enum": ["strData"]}},
    "required": ["messagetype"],
    "additionalProperties": False,
}

SELDON_BIN_SCHEMA = {
    "type": "object",
    "properties": {"messagetype": {"type": "string", "enum": ["binData"]}},
    "required": ["messagetype"],
    "additionalProperties": False,
}


SELDON_CUSTOM_DEFINITION = {
    "type": "object",
    "properties": {
        "messagetype": {
            "type": "string",
            "not": {
                "enum": [
                    "tensor",
                    "ndarray",
                    "tftensor",
                    "jsonData",
                    "strData",
                    "binData",
                ]
            },
        },
        "schema": {"type": "object"},
    },
    "required": ["messagetype"],
    "additionalProperties": False,
}


TENSOR_DATA_TYPES = [
    "BOOL",
    "UINT8",
    "UINT16",
    "UINT32",
    "UINT64",
    "INT8",
    "INT16",
    "INT32",
    "INT64",
    "FP16",
    "FP32",
    "FP64",
    "BYTES",
]


METADATA_TENSOR_SCHEMA = {
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "datatype": {"type": "string", "enum": TENSOR_DATA_TYPES},
        "shape": {"type": "array", "items": {"type": "integer"}},
    },
    "additionalProperties": False,
}


INPUTS_OUTPUTS_SCHEMA = {
    "type": "array",
    "items": {
        "anyOf": [
            SELDON_ARRAY_SCHEMA,
            SELDON_JSON_SCHEMA,
            SELDON_STR_SCHEMA,
            SELDON_BIN_SCHEMA,
            SELDON_CUSTOM_DEFINITION,
            METADATA_TENSOR_SCHEMA,
        ]
    },
    "additionalProperties": False,
}


JSON_SCHEMA = {
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "versions": {"type": "array", "items": {"type": "string"}},
        "platform": {"type": "string"},
        "inputs": INPUTS_OUTPUTS_SCHEMA,
        "outputs": INPUTS_OUTPUTS_SCHEMA,
        "custom": {"type": "object", "additionalProperties": {"type": "string"}},
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
    SeldonInvalidMetadataError: if data cannot be properly validated

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
        "name": image_name,
        "versions": [image_version],
        "platform": "",
        "inputs": [],
        "outputs": [],
        "custom": {},
    }

    data = {**default_meta, **data}

    try:
        validate(data, JSON_SCHEMA)
    except ValidationError as e:
        raise SeldonInvalidMetadataError(e)

    logger.debug(f"Successfully validated metadata:\n{json.dumps(data, indent=2)}")
    return data
