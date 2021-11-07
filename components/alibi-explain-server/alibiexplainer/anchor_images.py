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
# alibiexplainer/alibiexplainer/anchor_images.py
# and then modified.
#

import logging
import numpy as np
import alibi
from alibi.api.interfaces import Explanation
from alibi.utils.wrappers import ArgmaxTransformer
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.constants import SELDON_LOGLEVEL
from typing import Callable, List, Optional

logging.basicConfig(level=SELDON_LOGLEVEL)


class AnchorImages(ExplainerWrapper):
    def __init__(
        self,
        explainer: Optional[alibi.explainers.AnchorImage],
        **kwargs
    ) -> None:
        if explainer is None:
            raise Exception("Anchor images requires a built explainer")
        self.anchors_image = explainer
        self.kwargs = kwargs

    def explain(self, inputs: List) -> Explanation:
        arr = np.array(inputs)
        logging.info("Calling explain on image of shape %s", (arr.shape,))
        logging.info("anchor image call with %s", self.kwargs)
        anchor_exp = self.anchors_image.explain(arr[0], **self.kwargs)
        return anchor_exp
