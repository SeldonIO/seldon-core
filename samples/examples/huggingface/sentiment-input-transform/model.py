# Copyright (c) 2024 Seldon Technologies Ltd.

# Use of this software is governed BY
# (1) the license included in the LICENSE file or
# (2) if the license included in the LICENSE file is the Business Source License 1.1,
# the Change License after the Change Date as each is defined in accordance with the LICENSE file.

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
