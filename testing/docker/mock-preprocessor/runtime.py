from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse, ResponseOutput


class MockPreprocessor(MLModel):
    """
    Mock runtime which returns the features for the next model on the graph.
    """

    async def predict(self, inference_request: InferenceRequest) -> InferenceResponse:
        return InferenceResponse(
            model_name=self.name,
            model_version=self.version,
            outputs=[
                ResponseOutput(
                    name="input-0",
                    shape=[1, 4],
                    datatype="FP32",
                    data=[0.1, 0.2, 0.3, 0.4],
                )
            ],
        )
