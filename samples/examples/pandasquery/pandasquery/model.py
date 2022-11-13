# Copyright 2022 Seldon Technologies Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse
from mlserver.codecs import PandasCodec
from mlserver.errors import MLServerError
import pandas as pd
from fastapi import status

QUERY_KEY = "query"


class ModelParametersMissing(MLServerError):
  def __init__(self, model_name: str, reason: str):
    super().__init__(
      f"Parameters missing for model {model_name} {reason}", status.HTTP_400_BAD_REQUEST
    )

class PandasQueryRuntime(MLModel):

  async def load(self) -> bool:
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
