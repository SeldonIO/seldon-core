from alibi_detect.base import BaseDetector, concept_drift_dict
import numpy as np
from unittest import TestCase
from adserver.cd_model import AlibiDetectConceptDriftModel
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
        cd["data"]["distance"] = [0.1, 0.2, 0.3]
        cd["data"]["p_val"] = [0.1, 0.2, 0.3]
        cd["data"]["threshold"] = 0.1
        return cd

class TestDummyCDModel(TestCase):
    def test_basic(self):
        model = DummyCDModel()
        ad_model = AlibiDetectConceptDriftModel(
            "name", "s3://model", model=model, drift_batch_size=1
        )
        req = [[1, 2]]
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(res.data["data"]["is_drift"], 0)
        print(res.metrics)
        self.assertEqual(len(res.metrics), 8)

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
        self.assertEqual(len(res.metrics), 8)

        res = ad_model.process_event(req, headers)
        self.assertEqual(res, None)

class TestTextDriftModel(TestCase):

    def test_basic(self):
        model = DummyCDModel()
        ad_model = AlibiDetectConceptDriftModel(
            "imdb_text_drift", "gs://seldon-models/alibi-detect/cd/ks/imdb-0_4_4", drift_batch_size=2
        )
        req = ["This movie is NOT the same as the 1954 version with Judy Garland and James Mason, and that is a shame because the 1954 version is, in my opinion, much better. I am not denying Barbra Streisand's talent at all. She is a good actress and brilliant singer. I am not acquainted with Kris Kristofferson's other work and therefore I can't pass judgment on it. However, this movie leaves much to be desired. It is paced slowly, it has gratuitous nudity and foul language, and can be very difficult to sit through.<br /><br />However, I am not a big fan of rock music, so it's only natural that I would like the Judy Garland version better. See the 1976 film with Barbra and Kris, and judge for yourself."]
        headers = {}
        ad_model.load()
        res = ad_model.process_event(req, headers)
        res = ad_model.process_event(req, headers)
        self.assertEqual(res.data["data"]["is_drift"], 0)
