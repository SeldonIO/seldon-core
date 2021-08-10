import json
from typing import List, Dict, Optional, Union
import logging
import dill
import os
from adserver.constants import (
    REQUEST_ID_HEADER_NAME,
    NAMESPACE_HEADER_NAME,
)

from adserver.base import CEModel, ModelResponse
from adserver.base.storage import download_model
from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.metrics import DEFAULT_LABELS
from seldon_core.env_utils import NONIMPLEMENTED_MSG
from elasticsearch import Elasticsearch
from elasticsearch.exceptions import NotFoundError

SELDON_DEPLOYMENT_ID = DEFAULT_LABELS["seldon_deployment_name"]
SELDON_MODEL_ID = DEFAULT_LABELS["model_name"]
SELDON_PREDICTOR_ID = DEFAULT_LABELS["predictor_name"]


def _load_class_module(module_path: str) -> str:
    components = module_path.split(".")
    mod = __import__(".".join(components[:-1]))
    for comp in components[1:]:
        print(mod, comp)
        mod = getattr(mod, comp)
    return mod


class CustomMetricsModel(CEModel):  # pylint:disable=c-extension-no-member
    def __init__(
        self, name: str, storage_uri: str, elasticsearch_uri: str = None, model=None
    ):
        """
        Custom Metrics Model

        Parameters
        ----------
        name
             The name of the model
        storage_uri
             The URI location of the model
        """
        super().__init__(name)
        self.name = name
        self.storage_uri = storage_uri
        self.model = model
        self.ready = False
        self.elasticsearch_client = None

        if elasticsearch_uri:
            if NONIMPLEMENTED_MSG in [
                SELDON_DEPLOYMENT_ID,
                SELDON_MODEL_ID,
                SELDON_PREDICTOR_ID,
            ]:
                logging.error(
                    f"Elasticsearch URI provided but DEFAULT_LABELS not provided: {DEFAULT_LABELS}"
                )
            else:
                self.elasticsearch_client = Elasticsearch(elasticsearch_uri,verify_certs=False)

    def load(self):
        """
        Load the model from storage

        """
        if "/" in self.storage_uri:
            model_folder = download_model(self.storage_uri)
            self.model = dill.load(
                open(os.path.join(model_folder, "meta.pickle"), "rb")
            )
        else:
            # Load from locally available models
            MetricsClass = _load_class_module(self.storage_uri)
            self.model = MetricsClass()

        self.ready = True

    def process_event(self, inputs: Union[List, Dict], headers: Dict) -> Optional[ModelResponse]:
        """
        Process the event and return Alibi Detect score

        Parameters
        ----------
        inputs
             Input data
        headers
             Header options

        Returns
        -------
             SeldonResponse response

        """
        logging.info("PROCESSING Feedback Event.")
        logging.info(str(headers))
        logging.info("----")

        metrics: List[Dict] = []
        output: Dict = {}
        truth = None
        response = None
        error = None

        if not isinstance(inputs, dict):
            raise SeldonMicroserviceException(
                f"Data is not a dict: {json.dumps(inputs)}",
                status_code=400,
                reason="BAD_DATA",
            )

        if "truth" not in inputs:
            raise SeldonMicroserviceException(
                f"No truth value provided in: {json.dumps(inputs)}",
                status_code=400,
                reason="NO_TRUTH_VALUE",
            )
        else:
            truth = inputs["truth"]

        # We automatically add any metrics provided in the incoming request
        if "metrics" in inputs:
            metrics.extend(inputs["metrics"])

        # If response is provided then we can perform a comparison
        if "response" in inputs:
            response = inputs["response"]
        elif REQUEST_ID_HEADER_NAME in headers:
            # Otherwise if UUID is provided we can fetch from elasticsearch
            if not self.elasticsearch_client:
                error = "Seldon-Puid provided but elasticsearch client not configured"
            else:
                try:
                    seldon_puid = headers.get(REQUEST_ID_HEADER_NAME, "")
                    seldon_namespace = headers.get(NAMESPACE_HEADER_NAME, "")

                    # Currently only supports SELDON inference type (not kfserving)
                    elasticsearch_index = f"inference-log-seldon-{seldon_namespace}-{SELDON_DEPLOYMENT_ID}-{SELDON_PREDICTOR_ID}"

                    doc = self.elasticsearch_client.get(
                        index=elasticsearch_index, id=seldon_puid
                    )
                    response = (
                        doc.get("_source", {})
                        .get("response", None)
                        .get("instance", None)
                    )
                    if not response:
                        error = f"Elasticsearch index {elasticsearch_index} with id {seldon_puid} did not contain response value"
                except NotFoundError:
                    error = f"Elasticsearch index {elasticsearch_index} with id {seldon_puid} not found"
        else:
            error = "Neither response nor request Puid provided in headers"

        if error:
            raise SeldonMicroserviceException(
                error, status_code=400, reason="METRICS_SERVER_ERROR"
            )

        logging.error(f"{truth}, {response}")
        metrics_transformed = self.model.transform(truth, response)

        metrics.extend(metrics_transformed.metrics)

        return ModelResponse(data={}, metrics=metrics)
