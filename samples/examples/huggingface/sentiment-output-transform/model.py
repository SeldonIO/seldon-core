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
    res_list = self.decode_request(payload, default_codec=StringRequestCodec)
    scores = []
    for res in res_list:
      logger.debug("decoded data: %s",res)
      sentiment = json.loads(res)
      if sentiment["label"] == "POSITIVE":
        scores.append(1)
      else:
        scores.append(0)
    return NumpyRequestCodec.encode_response(
      model_name="sentiments",
      payload=np.array(scores)
    )
