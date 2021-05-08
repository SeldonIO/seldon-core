from alibi_detect.base import BaseDetector, outlier_prediction_dict
import numpy as np
from unittest import TestCase
from adserver.od_model import AlibiDetectOutlierModel
from adserver.constants import HEADER_RETURN_INSTANCE_SCORE
from typing import Dict
from adserver.base import ModelResponse


class DummyODModel(BaseDetector):
    def __init__(
        self,
        expected_return_instance_score: bool = False,
        expected_return_feature_score: bool = False,
        expect_return_is_outlier: int = 0,
    ):
        super().__init__()
        self.expected_return_instance_score = expected_return_instance_score
        self.expected_return_feature_score = expected_return_feature_score
        self.expect_return_is_outlier = expect_return_is_outlier

    def score(self, X: np.ndarray):
        pass

    def predict(
        self,
        X: np.ndarray,
        outlier_type: str = "instance",
        outlier_perc: float = 100.0,
        batch_size: int = int(1e10),
        return_feature_score: bool = True,
        return_instance_score: bool = True,
    ) -> Dict[Dict[str, str], Dict[np.ndarray, np.ndarray]]:
        assert return_instance_score == self.expected_return_instance_score
        assert return_feature_score == self.expected_return_feature_score
        od = outlier_prediction_dict()
        od["data"]["is_outlier"] = self.expect_return_is_outlier
        return od


class TestODModel(TestCase):
    def test_basic(self):
        model = DummyODModel()
        od_model = AlibiDetectOutlierModel("name", "s3://model", model=model)
        req = [1, 2]
        headers = {}
        res: ModelResponse = od_model.process_event(req, headers)
        self.assert_(res is not None)
        self.assertEqual(res.data["data"]["is_outlier"], 0)

    def test_no_return_instance_score(self):
        expect_return_is_outlier = 1
        model = DummyODModel(
            expected_return_instance_score=False,
            expect_return_is_outlier=expect_return_is_outlier,
        )
        ad_model = AlibiDetectOutlierModel("name", "s3://model", model=model)
        req = [1, 2]
        headers = {HEADER_RETURN_INSTANCE_SCORE: "false"}
        res: ModelResponse = ad_model.process_event(req, headers)
        self.assert_(res is not None)
        self.assertEqual(res.data["data"]["is_outlier"], expect_return_is_outlier)
