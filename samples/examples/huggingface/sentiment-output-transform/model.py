# Copyright (c) 2024 Seldon Technologies Ltd.

# Use of this software is governed BY
# (1) the license included in the LICENSE file or
# (2) if the license included in the LICENSE file is the Business Source License 1.1,
# the Change License after the Change Date as each is defined in accordance with the LICENSE file.

from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse, ResponseOutput
from mlserver.codecs import StringCodec, Base64Codec, NumpyRequestCodec
from mlserver.codecs.string import StringRequestCodec
from mlserver.codecs.numpy import NumpyRequestCodec
import base64
from mlserver.logging import logger
import numpy as np
import json

class SentimentOutputTransformRuntime(MLModel):

  async def load(self) -> bool:
    return self.ready

  async def predict(self, payload: InferenceRequest) -> InferenceResponse:
    logger.info("payload (output-transform): %s",payload)
    res_list = self.decode_request(payload, default_codec=StringRequestCodec)
    logger.info("res list (output-transform): %s",res_list)
    scores = []
    for res in res_list:
      logger.debug("decoded data (output transform): %s",res)
      #sentiment = json.loads(res)
      sentiment = res
      if sentiment["label"] == "POSITIVE":
        scores.append(1)
      else:
        scores.append(0)
    response =  NumpyRequestCodec.encode_response(
      model_name="sentiments",
      payload=np.array(scores)
    )
    logger.info("response (output-transform): %s", response)
    return response
