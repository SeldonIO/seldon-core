from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse, ResponseOutput
from mlserver.codecs.string import StringRequestCodec
from mlserver.logging import logger
import json


class SentimentInputTransformRuntime(MLModel):

  async def load(self) -> bool:
    return self.ready

  async def predict(self, payload: InferenceRequest) -> InferenceResponse:
    res_list = self.decode_request(payload, default_codec=StringRequestCodec)
    texts = []
    for res in res_list:
      logger.debug("decoded data: %s", res)
      text = json.loads(res)
      texts.append(text["text"])

    return StringRequestCodec.encode_response(
      model_name="sentiment",
      payload=texts
    )
