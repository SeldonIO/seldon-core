# Copyright 2019 kubeflow.org.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Original source from https://github.com/kubeflow/kfserving/blob/master/python/
# alibiexplainer/alibiexplainer/anchor_tabular.py
# and since modified
#

import logging
from typing import List, Optional

import alibi
import numpy as np
from alibi.api.interfaces import Explanation

from alibiexplainer.constants import (
    EXPLAIN_RANDOM_SEED,
    EXPLAIN_RANDOM_SEED_VALUE,
    SELDON_LOGLEVEL,
)
from alibiexplainer.explainer_wrapper import ExplainerWrapper

logging.basicConfig(level=SELDON_LOGLEVEL)


class AnchorTabular(ExplainerWrapper):
    def __init__(self, explainer=Optional[alibi.explainers.AnchorTabular], **kwargs):
        if explainer is None:
            raise Exception("Anchor tabular requires a built explainer")
        self.anchors_tabular: alibi.explainers.AnchorTabular = explainer
        self.anchors_tabular = explainer
        if EXPLAIN_RANDOM_SEED == "True" and str(EXPLAIN_RANDOM_SEED_VALUE).isdigit():
            self.seed = int(EXPLAIN_RANDOM_SEED_VALUE)
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        if hasattr(self, "seed"):
            np.random.seed(self.seed)
        arr = np.array(inputs)
        # We assume the input has batch dimension
        # but Alibi explainers presently assume no batch
        anchor_exp = self.anchors_tabular.explain(arr[0], **self.kwargs)
        return anchor_exp
