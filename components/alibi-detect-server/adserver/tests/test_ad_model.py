from alibi_detect.base import BaseDetector, adversarial_prediction_dict
import numpy as np
from unittest import TestCase
from adserver.ad_model import AlibiDetectAdversarialDetectionModel
from adserver.constants import HEADER_RETURN_INSTANCE_SCORE
from typing import Dict


class DummyAEModel(BaseDetector):
    def __init__(
        self,
        expected_return_instance_score: bool = True,
        expected_is_adversarial: int = 1,
    ):
        super().__init__()
        self.expected_return_instance_score = expected_return_instance_score
        self.expected_is_adversarial = expected_is_adversarial

    def score(self, X: np.ndarray):
        pass

    def predict(
        self,
        X: np.ndarray,
        batch_size: int = int(1e10),
        return_instance_score: bool = True,
    ) -> Dict[Dict[str, str], Dict[str, np.ndarray]]:
        assert return_instance_score == self.expected_return_instance_score
        ad = adversarial_prediction_dict()
        ad["data"]["is_adversarial"] = self.expected_is_adversarial
        return ad


class TestAEModel(TestCase):
    def test_basic(self):
        model = DummyAEModel()
        ad_model = AlibiDetectAdversarialDetectionModel(
            "name", "s3://model", model=model
        )
        req = [1, 2]
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assert_(res is not None)
        self.assertEqual(res.data["data"]["is_adversarial"], 1)

    def test_no_return_instance_score(self):
        expected_is_adversarial = 1
        model = DummyAEModel(
            expected_return_instance_score=False,
            expected_is_adversarial=expected_is_adversarial,
        )
        ad_model = AlibiDetectAdversarialDetectionModel(
            "name", "s3://model", model=model
        )
        req = [1, 2]
        headers = {HEADER_RETURN_INSTANCE_SCORE: "false"}
        res = ad_model.process_event(req, headers)
        self.assert_(res is not None)
        self.assertEqual(res.data["data"]["is_adversarial"], expected_is_adversarial)
