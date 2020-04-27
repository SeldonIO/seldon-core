import os
import logging

from typing import Union, List, Dict

from seldon_core.metrics import split_image_tag


logger = logging.getLogger(__name__)

MODEL_IMAGE = os.environ.get("PREDICTIVE_UNIT_IMAGE")


class SeldonInvalidMetadataError(Exception):
    pass


class MetadataTensorValidator:
    """MetadataTensorValidator class

    Stores and validates metadata_tensor defined as follows

        $metadata_tensor =
        {
          "name" : $string,
          "datatype" : $string,
          "shape" : [ $number, ... ]
        }

    """

    def __init__(self, name: str, datatype: str, shape: Union[List[int]]):
        self.name = str(name)
        self.datatype = str(datatype)
        self.shape = self.validate_shape(shape)

    @staticmethod
    def validate_shape(shape: Union[List[int]]) -> Union[List[int]]:
        err_msg = "MetadataTensorValidator shape field must be a sequence of integers"
        if not isinstance(shape, (list, tuple)):
            logger.error(err_msg)
            raise SeldonInvalidMetadataError(err_msg)

        if not all(isinstance(number, int) for number in shape):
            logger.error(err_msg)
            raise SeldonInvalidMetadataError(err_msg)

        return shape

    def to_dict(self) -> Dict:
        return {
            "name": self.name,
            "datatype": self.datatype,
            "shape": self.shape,
        }


class ModelMetadataValidator:
    """ModelMetadataValidator class

    Stores and validates metadata_model_response defined as follows:

    $metadata_model_response =
        {
          "name" : $string,
          "versions" : [ $string, ... ] #optional,
          "platform" : $string,
          "inputs" : [ $metadata_tensor, ... ],
          "outputs" : [ $metadata_tensor, ... ]
        }
    """

    def __init__(
        self,
        name: str = None,
        versions: List[str] = None,
        platform: str = None,
        inputs: List[MetadataTensorValidator] = None,
        outputs: List[MetadataTensorValidator] = None,
    ):
        self.name = str(name) if name is not None else ""
        self.versions = self.validate_versions(versions)
        self.platform = str(platform) if platform is not None else ""
        self.inputs = self.validate_tensors(inputs)
        self.outputs = self.validate_tensors(outputs)

        logger.debug(f"Successfully validated ModelMetadataValidator: {self}")

    @staticmethod
    def validate_versions(versions: List[str],) -> List[str]:
        err_msg = "ModelMetadataValidator versions field must be a sequence of strings"
        if versions is None:
            return []
        if not isinstance(versions, (list, tuple)):
            raise SeldonInvalidMetadataError(err_msg)
        if not all(isinstance(v, str) for v in versions):
            logger.error(err_msg)
            raise SeldonInvalidMetadataError(err_msg)
        return versions

    @staticmethod
    def validate_tensors(
        tensors: List[MetadataTensorValidator],
    ) -> List[MetadataTensorValidator]:
        err_msg = (
            "ModelMetadataValidator inputs and outputs must be "
            "sequence of MetadataTensorValidators."
        )
        if tensors is None:
            return []
        if not isinstance(tensors, (list, tuple)):
            raise SeldonInvalidMetadataError(err_msg)
        if not all(isinstance(v, MetadataTensorValidator) for v in tensors):
            logger.error(err_msg)
            raise SeldonInvalidMetadataError(err_msg)
        return tensors

    def __repr__(self) -> str:
        return str(self)

    def __str__(self) -> str:
        return str(self.to_dict())

    def to_dict(self) -> Dict:
        return {
            "name": self.name,
            "versions": self.versions,
            "platform": self.platform,
            "inputs": [x.to_dict() for x in self.inputs],
            "outputs": [x.to_dict() for x in self.outputs],
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

    Read data from json and create ModelMetadataValidator and MetadataTensorValidator objects.
    This function reads data in safe manner from json: validation and exceptions
    will happen in ModelMetadataValidator and MetadataTensorValidator classes.

    SeldonInvalidMetadataError has been chosen for exception as validation mostly depend on having
    a correct type for specific components.
    """
    if MODEL_IMAGE is not None:
        image_name, image_version = split_image_tag(MODEL_IMAGE)
    else:
        image_name, image_version = "", ""
    name = data.get("name", image_name)
    versions = data.get("versions", [image_version])
    platform = data.get("platform", "")

    try:
        inputs = [
            MetadataTensorValidator(x.get("name"), x.get("datatype"), x.get("shape"))
            for x in data.get("inputs", [])
        ]

        outputs = [
            MetadataTensorValidator(x.get("name"), x.get("datatype"), x.get("shape"))
            for x in data.get("outputs", [])
        ]
    except (AttributeError, TypeError):
        raise SeldonInvalidMetadataError(
            "Model metadata inputs and outputs must be sequence of dictionaries."
        )

    meta = ModelMetadataValidator(
        name=name, versions=versions, platform=platform, inputs=inputs, outputs=outputs,
    )

    return meta.to_dict()
