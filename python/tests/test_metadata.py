import logging
import yaml
import json
import pytest

from unittest.mock import patch

from seldon_core.metrics import SeldonMetrics
from seldon_core.wrapper import get_rest_microservice
from seldon_core.metadata import (
    SeldonInvalidMetadataError,
    validate_model_metadata,
)

from jsonschema.exceptions import ValidationError
import os

# test default values


def test_default_values():
    meta = validate_model_metadata({})
    assert meta == {
        "apiVersion": "v2",
        "name": "",
        "versions": [""],
        "platform": "",
        "inputs": [],
        "outputs": [],
    }


def test_default_values_with_env():
    with patch("seldon_core.metadata.MODEL_IMAGE", "seldonio/sklearn-iris:0.1"):
        meta = validate_model_metadata({})
    assert meta == {
        "apiVersion": "v2",
        "name": "seldonio/sklearn-iris",
        "versions": ["0.1"],
        "platform": "",
        "inputs": [],
        "outputs": [],
    }


def test_default_values_with_colon_in_env():
    with patch("seldon_core.metadata.MODEL_IMAGE", "localhost:32000/sklearn-iris:0.1"):
        meta = validate_model_metadata({})
    assert meta == {
        "apiVersion": "v2",
        "name": "localhost:32000/sklearn-iris",
        "versions": ["0.1"],
        "platform": "",
        "inputs": [],
        "outputs": [],
    }


# V1 meta tests


def test_v1_array():
    data = """
        apiVersion: v1
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
          datatype: array
          shape: [ 2, 2 ]
        outputs:
          datatype: array
          shape: [ 1 ]
    """
    validate_model_metadata(yaml.safe_load(data))


def test_v1_json():
    data = """
        apiVersion: v1
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
          datatype: jsonData
        outputs:
          datatype: array
          shape: [ 1 ]
    """
    validate_model_metadata(yaml.safe_load(data))


def test_v1_json_with_schema():
    data = """
        apiVersion: v1
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
          datatype: jsonData
          schema:
              type: object
              properties:
                  names:
                      type: array
                      items:
                          type: string
                  data:
                    type: array
                    items:
                        type: number
                        format: double
        outputs:
          datatype: array
          shape: [ 1 ]
    """
    validate_model_metadata(yaml.safe_load(data))


def test_v1_str_data():
    data = """
        apiVersion: v1
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
          datatype: strData
        outputs:
          datatype: array
          shape: [ 1 ]
    """
    validate_model_metadata(yaml.safe_load(data))


def test_v1_bin_data():
    data = """
        apiVersion: v1
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
          datatype: binData
        outputs:
          datatype: array
          shape: [ 1 ]
    """
    validate_model_metadata(yaml.safe_load(data))


# V1 meta tests (failures)


@pytest.mark.parametrize(
    "invalid_input",
    [
        {},  # no such schema
        {"datatype": "mytype"},  # to such valid schema
        {"datatype": "array"},  # fails because shape is missing
        {"datatype": "array", "shape": "1, 2"},  # shape is wrong type
        {"datatype": "array", "shape": "1, 2"},  # shape is wrong type
        {"datatype": "array", "shape": [2, 2], "invalid": "field"},
        {"datatype": "jsonData", "invalid": "field"},
        {"datatype": "jsonData", "schema": "some string"},  # schema should be dict
        {"datatype": "strData", "invalid": "field"},
        {"datatype": "binData", "invalid": "field"},
    ],
)
def test_v1_invalid_inputs(invalid_input):
    valid_base = {
        "apiVersion": "v1",
        "inputs": {"datatype": "strData"},
        "outputs": {"datatype": "strData"},
    }
    with pytest.raises(SeldonInvalidMetadataError):
        validate_model_metadata({**valid_base, **{"inputs": invalid_input}})
        validate_model_metadata({**valid_base, **{"outputs": invalid_input}})


# V2 meta tests


def test_v2():
    data = """
        apiVersion: v2
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
        - datatype: BYTES
          name: input
          shape: [ 1, 4 ]
        outputs:
        - datatype: BYTES
          name: output
          shape: [ 3 ]
    """
    validate_model_metadata(yaml.safe_load(data))


# validate_model_metadata tests block


def test_validate_model_metadata():
    meta = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
    }
    with patch("seldon_core.metadata.MODEL_IMAGE", None):
        assert {"apiVersion": "v2", **meta} == validate_model_metadata(meta)


def test_validate_model_metadata_with_env():
    meta = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
    }
    with patch("seldon_core.metadata.MODEL_IMAGE", "seldonio/sklearn-iris:0.1"):
        assert {"apiVersion": "v2", **meta} == validate_model_metadata(meta)


def test_validate_model_metadata_with_colon_in_env():
    meta = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
    }
    with patch("seldon_core.metadata.MODEL_IMAGE", "localhost:32000/sklearn-iris:0.1"):
        assert {"apiVersion": "v2", **meta} == validate_model_metadata(meta)


@pytest.mark.parametrize("invalid_versions", ["v1", [1], "[v]", "[1]", 1, 1.1])
def test_validate_model_metadata_wrong_versions(invalid_versions):
    with pytest.raises(SeldonInvalidMetadataError):
        validate_model_metadata({"versions": invalid_versions})


@pytest.mark.parametrize(
    "invalid_tensor",
    [
        {"name": "tensor-name", "datatype": "data-type", "shape": [1, 2]},
        "some string",
        ["some string in array"],
    ],
)
def test_validate_model_metadata_wrong_inputs_outputs(invalid_tensor):
    # Note: this could not be combined with test_model_metadata_wrong_inputs_outputs as
    # second input there is a valid input here
    with pytest.raises(SeldonInvalidMetadataError):
        validate_model_metadata({"inputs": invalid_tensor})
    with pytest.raises(SeldonInvalidMetadataError):
        validate_model_metadata({"outputs": invalid_tensor})


# Microservice tests block

yaml_meta = """
---
name: test-name-env
versions: [ test-version ]
platform: test-platform
inputs:
- datatype: BYTES
  name: input
  shape: [ 1, 2 ]
outputs:
- datatype: BYTES
  name: output
  shape: [ 1, 2, 3 ]
"""


json_meta = """
{
    "name": "test_name_json",
    "versions": ["test-version"],
    "platform": "seldon",
    "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 2]}],
    "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 2, 3]}],
}
"""


class UserObject:
    METADATA_RESPONSE = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
    }

    def predict(self, X, features_names):
        logging.info("Predict called")
        return X

    def init_metadata(self):
        logging.info("init_metadata called")
        return self.METADATA_RESPONSE


def test_model_metadata_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert json.loads(rv.data)["data"]["ndarray"] == ["data"]

    rv = client.get("/metadata")
    assert rv.status_code == 200
    assert json.loads(rv.data) == {"apiVersion": "v2", **UserObject.METADATA_RESPONSE}


@pytest.mark.parametrize("env_value", [json_meta, yaml_meta,])
def test_model_metadata_value_in_env(env_value):
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()

    # it is enough to only patch call to get_rest_microservice as it caches the metadata
    with patch("os.environ", {"MODEL_METADATA": env_value}):
        app = get_rest_microservice(user_object, seldon_metrics)

    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert json.loads(rv.data)["data"]["ndarray"] == ["data"]

    rv = client.get("/metadata")
    assert rv.status_code == 200
    assert json.loads(rv.data) == {
        **UserObject.METADATA_RESPONSE,
        "apiVersion": "v2",
        **yaml.safe_load(env_value),
    }


def test_model_metadata_invalid_user_definition():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()

    # Break UserObject class to provide invalid init_metadata
    user_object.METADATA_RESPONSE["versions"] = "string-not-array-of-strings"

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert json.loads(rv.data)["data"]["ndarray"] == ["data"]

    rv = client.get("/metadata")
    assert rv.status_code == 500
    assert json.loads(rv.data) == {
        "status": {
            "code": -1,
            "info": "Model metadata unavailable",
            "reason": "MICROSERVICE_BAD_METADATA",
            "status": 1,
        }
    }
