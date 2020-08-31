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
# Originaly copied from https://github.com/kubeflow/kfserving/blob/master/python/alibiexplainer/alibiexplainer/explainer_wrapper.py
# and since modified
#

from typing import List, Dict, Optional


class ExplainerWrapper(object):

    def validate(self, training_data_url: Optional[str]):
        pass

    def explain(self, inputs: List) -> Dict:
        pass
