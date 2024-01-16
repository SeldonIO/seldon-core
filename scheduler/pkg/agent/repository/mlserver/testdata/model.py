# Copyright (c) 2024 Seldon Technologies Ltd.

# Use of this software is governed BY
# (1) the license included in the LICENSE file or
# (2) if the license included in the LICENSE file is the Business Source License 1.1,
# the Change License after the Change Date as each is defined in accordance with the LICENSE file.

from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse
from mlserver.codecs import PandasCodec
from mlserver.errors import MLServerError
import pandas as pd
from fastapi import status
from mlserver.logging import logger

QUERY_KEY = "query"


class ModelParametersMissing(MLServerError):
  def __init__(self, model_name: str, reason: str):
    super().__init__(
      f"Parameters missing for model {model_name} {reason}", status.HTTP_400_BAD_REQUEST
    )

class PandasQueryRuntime(MLModel):

  async def load(self) -> bool:
    logger.info("Loading with settings %s", self.settings)
    if self.settings.parameters is None or \
      self.settings.parameters.extra is None:
      raise ModelParametersMissing(self.name, "no settings.parameters.extra found")
    self.query = self.settings.parameters.extra[QUERY_KEY]
    if self.query is None:
      raise ModelParametersMissing(self.name, "no settings.parameters.extra.query found")
    self.ready = True

    return self.ready

  async def predict(self, payload: InferenceRequest) -> InferenceResponse:
    input_df: pd.DataFrame = PandasCodec.decode_request(payload)
    # run query on input_df and save in output_df
    output_df = input_df.query(self.query)
    if output_df.empty:
      output_df = pd.DataFrame({'status':["no rows satisfied " + self.query]})
    else:
      output_df["status"] = "row satisfied " + self.query
    return PandasCodec.encode_response(self.name, output_df, self.version)
