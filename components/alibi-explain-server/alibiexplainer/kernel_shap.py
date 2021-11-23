import logging
from typing import Callable, List, Optional

import alibi
import numpy as np
from alibi.api.interfaces import Explanation

from alibiexplainer.constants import SELDON_LOGLEVEL
from alibiexplainer.explainer_wrapper import ExplainerWrapper

logging.basicConfig(level=SELDON_LOGLEVEL)


class KernelShap(ExplainerWrapper):
    def __init__(self, explainer: Optional[alibi.explainers.KernelShap], **kwargs):
        if explainer is None:
            raise Exception("Kernel Shap requires a built explainer")
        self.kernel_shap = explainer
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        arr = np.array(inputs)
        logging.info("kernel Shap call with %s", self.kwargs)
        logging.info("kernel shap data shape %s", arr.shape)
        shap_exp = self.kernel_shap.explain(arr, l1_reg=False, **self.kwargs)
        return shap_exp
