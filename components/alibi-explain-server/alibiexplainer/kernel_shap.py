import logging
import numpy as np
import alibi
from alibi.api.interfaces import Explanation
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.constants import SELDON_LOGLEVEL
from typing import Callable, List, Optional

logging.basicConfig(level=SELDON_LOGLEVEL)


class KernelShap(ExplainerWrapper):
    def __init__(
        self,
        predict_fn: Callable,
        explainer: Optional[alibi.explainers.KernelShap],
        **kwargs
    ):
        if explainer is None:
            raise Exception("Kernel Shap requires a built explainer")
        self.predict_fn = predict_fn
        self.kernel_shap = explainer
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        arr = np.array(inputs)
        self.kernel_shap.predictor = self.predict_fn
        logging.info("kernel Shap call with %s", self.kwargs)
        logging.info("kernel shap data shape %s",arr.shape)
        shap_exp = self.kernel_shap.explain(arr, l1_reg=False, **self.kwargs)
        return shap_exp
