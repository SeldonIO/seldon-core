import logging
import numpy as np
import alibi
from alibi.api.interfaces import Explanation
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.constants import SELDON_LOGLEVEL
from typing import List, Optional

logging.basicConfig(level=SELDON_LOGLEVEL)


class TreeShap(ExplainerWrapper):
    def __init__(
        self,
        explainer: Optional[alibi.explainers.TreeShap],
        **kwargs
    ):
        if explainer is None:
            raise Exception("Tree Shap requires a built explainer")
        self.tree_shap = explainer
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        arr = np.array(inputs)
        logging.info("Tree Shap call with %s", self.kwargs)
        logging.info("kernel shap data shape %s",arr.shape)
        shap_exp = self.tree_shap.explain(arr, **self.kwargs)
        return shap_exp