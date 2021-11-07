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
# original source from https://github.com/kubeflow/kfserving/blob/master/python/
# alibiexplainer/alibiexplainer/anchor_text.py
# and since modified
#

import logging
import numpy as np
import spacy
import alibi
from alibi.api.interfaces import Explanation
from alibi.utils.download import spacy_model
from alibi.utils.wrappers import ArgmaxTransformer
from alibiexplainer.explainer_wrapper import ExplainerWrapper
from alibiexplainer.constants import SELDON_LOGLEVEL
from typing import Callable, List, Optional

logging.basicConfig(level=SELDON_LOGLEVEL)


class AnchorText(ExplainerWrapper):
    def __init__(
        self,
        predict_fn: Callable,
        explainer: Optional[alibi.explainers.AnchorText],
        spacy_language_model: str = "en_core_web_md",
        **kwargs
    ):
        self.predict_fn = predict_fn
        self.kwargs = kwargs
        logging.info("Anchor Text args %s", self.kwargs)
        if explainer is None:
            logging.info("Loading Spacy Language model for %s", spacy_language_model)
            spacy_model(model=spacy_language_model)
            self.nlp = spacy.load(spacy_language_model)
            logging.info("Language model loaded")
            self.anchors_text = alibi.explainers.AnchorText(
                predictor=predict_fn,
                sampling_strategy="unknown",
                nlp=self.nlp)
        else:
            self.anchors_text = explainer

    def explain(self, inputs: List) -> Explanation:
        anchor_exp = self.anchors_text.explain(inputs[0], **self.kwargs)
        return anchor_exp
