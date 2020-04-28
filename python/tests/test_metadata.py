import logging
import json
import pytest

from unittest.mock import patch

from seldon_core.metrics import SeldonMetrics
from seldon_core.wrapper import get_rest_microservice
from seldon_core.metadata import (
    SeldonInvalidMetadataError,
    MetadataTensorValidator,
    ModelMetadataValidator,
    validate_model_metadata,
)


# MetadataTensorValidator tests block


def test_metadata_tensor():
    mt = MetadataTensorValidator(name="tensor-name", datatype="data-type", shape=[1, 2])
    assert mt.to_dict() == {
        "name": "tensor-name",
        "datatype": "data-type",
        "shape": [1, 2],
    }


@pytest.mark.parametrize("invalid_shape", [1, [1.1], "1", "[1]", ["1"]])
def test_metadata_tensor_wrong_shape_type(invalid_shape):
    with pytest.raises(SeldonInvalidMetadataError):
        mt = MetadataTensorValidator(
            name="tensor-name", datatype="data-type", shape=invalid_shape
        )


# ModelMetadataValidator tests block


def test_model_metadata():
    meta = ModelMetadataValidator(
        name="model-name",
        versions=["model-version"],
        platform="platform-name",
        inputs=[
            MetadataTensorValidator(
                name="input-name", datatype="input-type", shape=[1, 2]
            )
        ],
        outputs=[
            MetadataTensorValidator(
                name="output-name", datatype="output-type", shape=[1, 2]
            )
        ],
    )
    assert meta.to_dict() == {
        "name": "model-name",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input-name", "datatype": "input-type", "shape": [1, 2]}],
        "outputs": [
            {"name": "output-name", "datatype": "output-type", "shape": [1, 2]}
        ],
    }


@pytest.mark.parametrize("invalid_versions", ["v1", [1], "[v]", "[1]", 1, 1.1])
def test_model_metadata_wrong_versions(invalid_versions):
    with pytest.raises(SeldonInvalidMetadataError):
        meta = ModelMetadataValidator(versions=invalid_versions)


@pytest.mark.parametrize(
    "invalid_tensor",
    [
        MetadataTensorValidator(name="tensor-name", datatype="data-type", shape=[1, 2]),
        [{"name": "tensor-name", "datatype": "data-type", "shape": [1, 2],}],
        {"name": "tensor-name", "datatype": "data-type", "shape": [1, 2],},
        "some string",
        ["some string in array"],
    ],
)
def test_model_metadata_wrong_inputs_outputs(invalid_tensor):
    with pytest.raises(SeldonInvalidMetadataError):
        meta = ModelMetadataValidator(inputs=invalid_tensor)
    with pytest.raises(SeldonInvalidMetadataError):
        meta = ModelMetadataValidator(outputs=invalid_tensor)


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
        assert meta == validate_model_metadata(meta)


def test_validate_model_metadata_with_env():
    meta = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
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
        MetadataTensorValidator(name="tensor-name", datatype="data-type", shape=[1, 2]),
        {"name": "tensor-name", "datatype": "data-type", "shape": [1, 2],},
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
    assert json.loads(rv.data) == UserObject.METADATA_RESPONSE


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
