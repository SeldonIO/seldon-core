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
from mlserver.types import InferenceRequest, InferenceResponse, ResponseOutput
from mlserver.codecs.string import StringRequestCodec
from mlserver.logging import logger
import json


class SentimentInputTransformRuntime(MLModel):

  async def load(self) -> bool:
    return self.ready

  async def predict(self, payload: InferenceRequest) -> InferenceResponse:
    logger.info("payload (input-transform): %s",payload)
    res_list = self.decode_request(payload, default_codec=StringRequestCodec)
    logger.info("res list (input-transform): %s",res_list)
    texts = []
    for res in res_list:
      logger.info("decoded data (input-transform): %s", res)
      #text = json.loads(res)
      text = res
      texts.append(text["text"])

    logger.info("transformed data (input-transform): %s", texts)
    response =  StringRequestCodec.encode_response(
      model_name="sentiment",
      payload=texts
    )
    logger.info("response (input-transform): %s", response)
    return response
