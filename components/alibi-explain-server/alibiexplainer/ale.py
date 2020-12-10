import logging
import numpy as np
import alibi
from alibi.api.interfaces import Explanation
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.constants import SELDON_LOGLEVEL
from typing import Callable, List, Optional

logging.basicConfig(level=SELDON_LOGLEVEL)


class ALE(ExplainerWrapper):
    def __init__(
        self,
        predict_fn: Callable,
        explainer: Optional[alibi.explainers.ale.ALE],
        **kwargs
    ):
        if explainer is None:
            raise Exception("ALE requires a built explainer")
        self.predict_fn = predict_fn
        self.ale = explainer
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        arr = np.array(inputs)
        self.ale.predictor = self.predict_fn
        logging.info("ALE call with %s", self.kwargs)
        logging.info("ALE data shape %s",arr.shape)
        ale_exp = self.ale.explain(arr, **self.kwargs)
        return ale_exp
