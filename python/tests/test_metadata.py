import logging
import yaml
import json
import pytest

from unittest.mock import patch

from seldon_core.metrics import SeldonMetrics
from seldon_core.proto import prediction_pb2
from seldon_core.wrapper import get_rest_microservice, SeldonModelGRPC
from seldon_core.metadata import SeldonInvalidMetadataError, validate_model_metadata
from seldon_core.utils import json_to_seldon_model_metadata

from google.protobuf import json_format


# test default values


def test_default_values():
    meta = validate_model_metadata({})
    assert meta == {
        "name": "",
        "versions": [""],
        "platform": "",
        "inputs": [],
        "outputs": [],
        "custom": {},
    }


def test_default_values_with_env():
    with patch("seldon_core.metadata.MODEL_IMAGE", "seldonio/sklearn-iris:0.1"):
        meta = validate_model_metadata({})
    assert meta == {
        "name": "seldonio/sklearn-iris",
        "versions": ["0.1"],
        "platform": "",
        "inputs": [],
        "outputs": [],
        "custom": {},
    }


def test_default_values_with_colon_in_env():
    with patch("seldon_core.metadata.MODEL_IMAGE", "localhost:32000/sklearn-iris:0.1"):
        meta = validate_model_metadata({})
    assert meta == {
        "name": "localhost:32000/sklearn-iris",
        "versions": ["0.1"],
        "platform": "",
        "inputs": [],
        "outputs": [],
        "custom": {},
    }


# V1 meta tests


@pytest.mark.parametrize(
    "messagetype",
    ["tensor", "ndarray", "tftensor", "jsonData", "binData", "strData", "custom_data"],
)
def test_v1_only_messagetype(messagetype):
    data = f"""
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
        - messagetype: {messagetype}
        outputs:
        - messagetype: {messagetype}
        custom:
          tag-key: tag-value
    """
    meta_json = validate_model_metadata(yaml.safe_load(data))
    meta_proto = json_to_seldon_model_metadata(meta_json)
    assert meta_json == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [{"messagetype": f"{messagetype}"}],
        "outputs": [{"messagetype": f"{messagetype}"}],
        "custom": {"tag-key": "tag-value"},
    }
    assert json.loads(json_format.MessageToJson(meta_proto)) == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [{"messagetype": f"{messagetype}"}],
        "outputs": [{"messagetype": f"{messagetype}"}],
        "custom": {"tag-key": "tag-value"},
    }


def test_v1_mixed_multiple_inputs():
    data = """
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
        - messagetype: "tensor"
          schema:
            names: [a, b, c, d]
            shape: [ 2, 2 ]
        - messagetype: jsonData
        outputs:
        - messagetype: "binData"
        custom:
          tag-key: tag-value
    """
    meta_json = validate_model_metadata(yaml.safe_load(data))
    meta_proto = json_to_seldon_model_metadata(meta_json)
    assert meta_json == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {
                "messagetype": "tensor",
                "schema": {"names": ["a", "b", "c", "d"], "shape": [2, 2]},
            },
            {"messagetype": "jsonData"},
        ],
        "outputs": [{"messagetype": "binData"}],
        "custom": {"tag-key": "tag-value"},
    }
    assert json.loads(json_format.MessageToJson(meta_proto)) == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {
                "messagetype": "tensor",
                "schema": {"names": ["a", "b", "c", "d"], "shape": [2, 2]},
            },
            {"messagetype": "jsonData"},
        ],
        "outputs": [{"messagetype": "binData"}],
        "custom": {"tag-key": "tag-value"},
    }


@pytest.mark.parametrize(
    "messagetype", ["tensor", "ndarray", "tftensor", "custom_data"]
)
def test_v1_array(messagetype):
    data = f"""
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
        - messagetype: {messagetype}
          schema:
            names: [a, b, c, d]
            shape: [ 2, 2 ]
        outputs:
        - messagetype: {messagetype}
          schema:
            shape: [ 1 ]
        custom:
          tag-key: tag-value
    """
    meta_json = validate_model_metadata(yaml.safe_load(data))
    meta_proto = json_to_seldon_model_metadata(meta_json)
    assert meta_json == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {
                "messagetype": f"{messagetype}",
                "schema": {"names": ["a", "b", "c", "d"], "shape": [2, 2]},
            }
        ],
        "outputs": [{"messagetype": f"{messagetype}", "schema": {"shape": [1]}}],
        "custom": {"tag-key": "tag-value"},
    }
    assert json.loads(json_format.MessageToJson(meta_proto)) == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {
                "messagetype": f"{messagetype}",
                "schema": {"names": ["a", "b", "c", "d"], "shape": [2.0, 2.0]},
            }
        ],
        "outputs": [{"messagetype": f"{messagetype}", "schema": {"shape": [1.0]}}],
        "custom": {"tag-key": "tag-value"},
    }


def test_v1_json_with_schema():
    data = """
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
        - messagetype: jsonData
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
        - messagetype: jsonData
        custom:
          tag-key: tag-value
    """
    meta_json = validate_model_metadata(yaml.safe_load(data))
    meta_proto = json_to_seldon_model_metadata(meta_json)
    assert meta_json == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {
                "messagetype": "jsonData",
                "schema": {
                    "type": "object",
                    "properties": {
                        "names": {"type": "array", "items": {"type": "string"}},
                        "data": {
                            "type": "array",
                            "items": {"type": "number", "format": "double"},
                        },
                    },
                },
            }
        ],
        "outputs": [{"messagetype": "jsonData"}],
        "custom": {"tag-key": "tag-value"},
    }
    assert json.loads(json_format.MessageToJson(meta_proto)) == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {
                "messagetype": "jsonData",
                "schema": {
                    "properties": {
                        "names": {"items": {"type": "string"}, "type": "array"},
                        "data": {
                            "type": "array",
                            "items": {"type": "number", "format": "double"},
                        },
                    },
                    "type": "object",
                },
            }
        ],
        "outputs": [{"messagetype": "jsonData"}],
        "custom": {"tag-key": "tag-value"},
    }


# V1 meta tests (failures)


@pytest.mark.parametrize(
    "invalid_input",
    [
        {},  # no such schema
        {"messagetype": "mytype"},  # to such valid schema
        {"messagetype": "array"},  # fails because shape is missing
        {"messagetype": "array", "shape": "1, 2"},  # shape is wrong type
        {"messagetype": "array", "shape": "1, 2"},  # shape is wrong type
        {"messagetype": "array", "shape": [2, 2], "invalid": "field"},
        {"messagetype": "jsonData", "invalid": "field"},
        {"messagetype": "jsonData", "schema": "some string"},  # schema should be dict
        {"messagetype": "strData", "invalid": "field"},
        {"messagetype": "binData", "invalid": "field"},
    ],
)
def test_v1_invalid_inputs(invalid_input):
    valid_base = {
        "inputs": [{"messagetype": "strData"}],
        "outputs": [{"messagetype": "strData"}],
    }
    with pytest.raises(SeldonInvalidMetadataError):
        validate_model_metadata({**valid_base, **{"inputs": invalid_input}})
        validate_model_metadata({**valid_base, **{"outputs": invalid_input}})


@pytest.mark.parametrize(
    "messagetype", ["tensor", "ndarray", "tftensor", "binData", "strData"]
)
def test_v1_invalid_schema_fields(messagetype):
    meta = {
        "inputs": [
            {"messagetype": messagetype, "schema": {"custom-field": "custom-def"}}
        ]
    }
    with pytest.raises(SeldonInvalidMetadataError):
        validate_model_metadata(meta)


@pytest.mark.parametrize("messagetype", ["jsonData", "customData"])
def test_v1_valid_custom_schema(messagetype):
    meta = {
        "inputs": [
            {"messagetype": messagetype, "schema": {"custom-field": "custom-def"}}
        ]
    }
    validate_model_metadata(meta)


# V2 meta tests


def test_v2():
    data = """
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
        custom:
          tag-key: tag-value
    """
    meta_json = validate_model_metadata(yaml.safe_load(data))
    meta_proto = json_to_seldon_model_metadata(meta_json)
    assert meta_json == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [{"datatype": "BYTES", "name": "input", "shape": [1, 4]}],
        "outputs": [{"datatype": "BYTES", "name": "output", "shape": [3]}],
        "custom": {"tag-key": "tag-value"},
    }
    assert json.loads(json_format.MessageToJson(meta_proto)) == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": ["1", "4"]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": ["3"]}],
        "custom": {"tag-key": "tag-value"},
    }


# mix test


def test_v1_v2_mix():
    data = """
        name: my-model-name
        versions: [ my-model-version-01 ]
        platform: seldon
        inputs:
        - datatype: BYTES
          name: input
          shape: [ 1, 4 ]
        - messagetype: jsonData
        outputs:
        - datatype: BYTES
          name: output
          shape: [ 3 ]
        custom:
          tag-key: tag-value
    """
    meta_json = validate_model_metadata(yaml.safe_load(data))
    meta_proto = json_to_seldon_model_metadata(meta_json)
    assert meta_json == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {"datatype": "BYTES", "name": "input", "shape": [1, 4]},
            {"messagetype": "jsonData"},
        ],
        "outputs": [{"datatype": "BYTES", "name": "output", "shape": [3]}],
        "custom": {"tag-key": "tag-value"},
    }
    assert json.loads(json_format.MessageToJson(meta_proto)) == {
        "name": "my-model-name",
        "versions": ["my-model-version-01"],
        "platform": "seldon",
        "inputs": [
            {"name": "input", "datatype": "BYTES", "shape": ["1", "4"]},
            {"messagetype": "jsonData"},
        ],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": ["3"]}],
        "custom": {"tag-key": "tag-value"},
    }


# validate_model_metadata tests block


def test_validate_model_metadata():
    meta = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
        "custom": {"tag-key": "tag-value"},
    }
    with patch("seldon_core.metadata.MODEL_IMAGE", None):
        assert meta == validate_model_metadata(meta)


def test_validate_model_metadata_with_env():
    meta = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
        "custom": {"tag-key": "tag-value"},
    }
    with patch("seldon_core.metadata.MODEL_IMAGE", "seldonio/sklearn-iris:0.1"):
        assert meta == validate_model_metadata(meta)


def test_validate_model_metadata_with_colon_in_env():
    meta = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
        "custom": {"tag-key": "tag-value"},
    }
    with patch("seldon_core.metadata.MODEL_IMAGE", "localhost:32000/sklearn-iris:0.1"):
        assert meta == validate_model_metadata(meta)


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
custom:
  tag-key: tag-value
"""


json_meta = """
{
    "name": "test_name_json",
    "versions": ["test-version"],
    "platform": "seldon",
    "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 2]}],
    "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 2, 3]}],
    "custom": {"tag-key": "tag-value"},
}
"""


class UserObject:
    METADATA_RESPONSE = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
        "custom": {"tag-key": "tag-value"},
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


def test_model_metadata_ok_grpc():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    resp = app.Metadata(None, None)
    assert json.loads(json_format.MessageToJson(resp)) == {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": ["1"]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": ["1"]}],
        "custom": {"tag-key": "tag-value"},
    }


@pytest.mark.parametrize("env_value", [json_meta, yaml_meta])
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
