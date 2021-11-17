import logging
from typing import Callable, List, Optional

import alibi
import numpy as np
from alibi.api.interfaces import Explanation

from alibiexplainer.constants import SELDON_LOGLEVEL
from alibiexplainer.explainer_wrapper import ExplainerWrapper

logging.basicConfig(level=SELDON_LOGLEVEL)


class ALE(ExplainerWrapper):
    def __init__(self, explainer: Optional[alibi.explainers.ale.ALE], **kwargs):
        if explainer is None:
            raise Exception("ALE requires a built explainer")
        self.ale = explainer
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        arr = np.array(inputs)
        logging.info("ALE call with %s", self.kwargs)
        logging.info("ALE data shape %s", arr.shape)
        ale_exp = self.ale.explain(arr, **self.kwargs)
        return ale_exp
