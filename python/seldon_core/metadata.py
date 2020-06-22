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
        "messagetype": {"type": "string", "enum": ["array"]},
        "names": {"type": "array", "items": {"type": "string"}},
        "shape": {"type": "array", "items": {"type": "integer"}},
    },
    "required": ["messagetype", "shape"],
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
        "datatype": {"type": "string", "enum": TENSOR_DATA_TYPES},
        "name": {"type": "string"},
        "shape": {"type": "array", "items": {"type": "integer"}},
    },
    "additionalProperties": False,
}


INPUTS_OUTPUTS_SCHEMA = {
    "type": "array",
    "items": {
        "oneOf": [
            SELDON_ARRAY_SCHEMA,
            SELDON_JSON_SCHEMA,
            SELDON_STR_SCHEMA,
            SELDON_BIN_SCHEMA,
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
    },
    "additionalProperties": False,
}



# SELDON_MESSAGE_SCHEMA = {
#     "type": "object",
#     "properties": {
#         "payloadtype": {
#             "type": "string",
#             "enum": ["data", "jsonData", "strData", "binData"],
#         },
#         "datatype": {"type": "string", "enum": TENSOR_DATA_TYPES},
#         "names": {"type": "array", "items": {"type": "string"}},
#         "shape": {"type": "array", "items": {"type": "integer"}},
#     },
# }



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
        "name": image_name,
        "versions": [image_version],
        "platform": "",
        "inputs": [],
        "outputs": [],
    }

    data = {**default_meta, **data}

    try:
        validate(data, JSON_SCHEMA)
    except ValidationError as e:
        raise SeldonInvalidMetadataError(e)

    logger.debug(f"Successfully validated metadata:\n{json.dumps(data, indent=2)}")
    return data
