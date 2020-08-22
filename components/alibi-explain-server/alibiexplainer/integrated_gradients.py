import logging
import numpy as np
import alibi
from alibi.api.interfaces import Explanation
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.constants import SELDON_LOGLEVEL
from typing import Callable, List, Optional
from tensorflow import keras

logging.basicConfig(level=SELDON_LOGLEVEL)


class IntegratedGradients(ExplainerWrapper):
    def __init__(
        self,
        keras_model: keras.Model,
        n_steps: int = 50,
        internal_batch_size: int = 100,
        method: str = "gausslegendre",
        layer: Optional[int] = None,
        **kwargs
    ):
        if keras_model is None:
            raise Exception("Integrated Gradients requires a Keras model")
        self.keras_model : keras.Model = keras_model
        keras_layer = None
        if layer is not None:
            keras_layer = keras_model.layers[layer]
        self.integrated_gradients: alibi.explainers.integrated_gradients = alibi.explainers.IntegratedGradients(keras_model,
                                 layer=keras_layer,
                                 n_steps=n_steps,
                                 method=method,
                                 internal_batch_size=internal_batch_size)
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        arr = np.array(inputs)
        print(arr.shape)
        logging.info("Integrated gradients call")
        predictions = self.keras_model(arr).numpy().argmax(axis=1)
        logging.info("predictions shape %s", predictions.shape)
        explanation = self.integrated_gradients.explain(arr,baselines=None, target=predictions)
        return explanation
