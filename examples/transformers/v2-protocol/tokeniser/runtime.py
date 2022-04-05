import os
from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse
from mlserver.codecs import NumpyCodec
from mlserver.codecs.string import StringRequestCodec, StringCodec
from transformers import GPT2Tokenizer

TOKENIZER_TYPE_ENV_NAME = "SELDON_TOKENIZER_TYPE"
TOKENIZER_TYPE_ENCODE = "ENCODER"

class Tokeniser(MLModel):
    async def load(self) -> bool:
        self._tokeniser = GPT2Tokenizer.from_pretrained("gpt2")
        self._tokenizer_type = os.environ.get(TOKENIZER_TYPE_ENV_NAME, TOKENIZER_TYPE_ENCODE)

        self.ready = True
        return self.ready

    async def predict(self, inference_request: InferenceRequest) -> InferenceResponse:
        outputs = None
        if self._tokenizer_type == TOKENIZER_TYPE_ENCODE:
            sentences = StringRequestCodec.decode(inference_request)
            tokenised = self._tokeniser(sentences, return_tensors="np")
            
            outputs = []
            for name, payload in tokenised.items():
                inference_output = NumpyCodec.encode(name=name, payload=payload)
                # Transformer's TF GPT2 model expects `INT32` inputs by default, so
                # let's enforce them
                inference_output.datatype = "INT32"
                outputs.append(inference_output)
        else:
            logits = NumpyCodec.decode(inference_request.inputs[0])
            # take the best next token probability of the last token of input ( greedy approach)
            next_token = logits.argmax(axis=2)[0]
            next_token_str = self._tokeniser.decode(
                next_token[-1:], skip_special_tokens=True, clean_up_tokenization_spaces=True
            ).strip()
            outputs = [StringCodec.encode("next_token", [next_token_str])]

        return InferenceResponse(
            model_name=self.name, model_version=self.version, outputs=outputs
        )
