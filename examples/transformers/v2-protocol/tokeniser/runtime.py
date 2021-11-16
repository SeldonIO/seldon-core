from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse
from mlserver.codecs import NumpyCodec
from mlserver.codecs.string import StringRequestCodec
from transformers import GPT2Tokenizer


class Tokeniser(MLModel):
    async def load(self) -> bool:
        self._tokeniser = GPT2Tokenizer.from_pretrained("gpt2")

        self.ready = True
        return self.ready

    async def predict(self, inference_request: InferenceRequest) -> InferenceResponse:
        sentences = StringRequestCodec.decode(inference_request)
        tokenised = self._tokeniser(sentences, return_tensors="np")

        outputs = []
        for name, payload in tokenised.items():
            inference_output = NumpyCodec.encode(name=name, payload=payload)
            # Transformer's TF GPT2 model expects `INT32` inputs by default, so
            # let's enforce them
            inference_output.datatype = "INT32"
            outputs.append(inference_output)

        return InferenceResponse(
            model_name=self.name, model_version=self.version, outputs=outputs
        )
