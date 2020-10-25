from alibi_detect.base import BaseDetector, concept_drift_dict
import numpy as np
from unittest import TestCase
from adserver.cd_model import AlibiDetectConceptDriftModel
from adserver.constants import HEADER_RETURN_INSTANCE_SCORE
from typing import Dict


class DummyCDModel(BaseDetector):
    def __init__(self, expect_return_is_drift: int = 0):
        super().__init__()
        self.expect_return_is_drift = expect_return_is_drift

    def score(self, X: np.ndarray):
        pass

    def predict(
        self, X: np.ndarray, drift_type: str = "batch", return_p_val: bool = True
    ) -> Dict[Dict[str, str], Dict[str, np.ndarray]]:
        cd = concept_drift_dict()
        cd["data"]["is_drift"] = self.expect_return_is_drift
        return cd


class TestAEModel(TestCase):
    def test_basic(self):
        model = DummyCDModel()
        ad_model = AlibiDetectConceptDriftModel(
            "name", "s3://model", model=model, drift_batch_size=1
        )
        req = [[1, 2]]
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(res.data["data"]["is_drift"], 0)

    def test_batch(self):
        model = DummyCDModel()
        ad_model = AlibiDetectConceptDriftModel(
            "name", "s3://model", model=model, drift_batch_size=2
        )
        req = [[1, 2]]
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(res, None)

        res = ad_model.process_event(req, headers)
        self.assertEqual(res.data["data"]["is_drift"], 0)

        res = ad_model.process_event(req, headers)
        self.assertEqual(res, None)
