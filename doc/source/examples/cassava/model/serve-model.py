from mlserver import MLModel
from mlserver.codecs import decode_args
import numpy as np
import tensorflow as tf
import tensorflow_hub as hub


# Define a class for our Model, inheriting the MLModel class from MLServer
class CassavaModel(MLModel):
    # Load the model into memory
    async def load(self) -> bool:
        tf.config.experimental.set_visible_devices([], "GPU")
        model_path = "."
        self._model = hub.KerasLayer(model_path)
        self.ready = True
        return self.ready

    # Logic for making predictions against our model
    @decode_args
    async def predict(self, payload: np.ndarray) -> np.ndarray:
        # convert payload to tf.tensor
        payload_tensor = tf.constant(payload)

        # Make predictions
        predictions = self._model(payload_tensor)
        predictions_max = tf.argmax(predictions, axis=-1)

        # convert predictions to np.ndarray
        response_data = np.array(predictions_max)

        return response_data
