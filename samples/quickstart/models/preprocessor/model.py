import re
import numpy as np

from mlserver import MLModel
from mlserver.codecs import NumpyCodec, StringCodec
from mlserver.types import InferenceRequest, InferenceResponse


class PreprocessorModel(MLModel):
    def __init__(self, settings):
        super().__init__(settings)

    async def load(self) -> bool:
        self.ready = True 
        return self.ready 
        
    async def predict(self, inference_request: InferenceRequest) -> InferenceResponse:

        input_text = inference_request.inputs[0]
        decoded_string = StringCodec.decode_input(input_text)[0]

        nums = extract_numerical_values(decoded_string)        
        nums_encoded = NumpyCodec.encode_output('output', nums)

        return InferenceResponse(
            model_name=self._settings.name,
            model_version=self._settings.version,
            outputs=[nums_encoded]
        )


## Extracts numerical values from a formatted text and outputs a vector of numerical values.
def extract_numerical_values(input_text):

    # Find key-value pairs in text
    pattern = r'"[^"]+":\s*"([^"]+)"'
    matches = re.findall(pattern, input_text)
    
    # Extract numerical values
    numerical_values = []
    for value in matches:
        cleaned_value = value.replace(",", "")
        if cleaned_value.isdigit():  # Integer
            numerical_values.append(int(cleaned_value))
        else:
            try:  
                numerical_values.append(float(cleaned_value))
            except ValueError:
                pass  
    
    # Convert to a numpy vector
    return np.array(numerical_values).reshape(1,-1)
