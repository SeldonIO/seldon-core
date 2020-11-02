from unittest import TestCase
from adserver.cm_model import CustomMetricsModel
from typing import Dict
import numpy as np


class TestAEModel(TestCase):
    def test_binary(self):
        ad_model = CustomMetricsModel(
            "name", "adserver.cm_models.binary_metrics.BinaryMetrics"
        )
        ad_model.load()
        req = {"truth": [0], "response": 1}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_false_positive")

        req = {"truth": [1], "response": 1}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

        req = {"truth": np.array([1]), "response": 1}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

    def test_multiclass_numeric(self):
        ad_model = CustomMetricsModel(
            "name", "adserver.cm_models.multiclass_numeric.MultiClassNumeric"
        )
        ad_model.load()
        req = {"truth": [4], "response": 8}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 2)
        self.assertTrue(
            res.metrics[0]["key"]
            in ["seldon_metric_false_positive", "seldon_metric_false_negative"]
        )
        self.assertTrue(
            res.metrics[1]["key"]
            in ["seldon_metric_false_positive", "seldon_metric_false_negative"]
        )

        req = {"truth": [7], "response": 7}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

        req = {"truth": np.array([7]), "response": 7}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

    def test_multiclass_onehot(self):
        ad_model = CustomMetricsModel(
            "name", "adserver.cm_models.multiclass_one_hot.MulticlassOneHot"
        )
        ad_model.load()

        req = {"truth": [0, 0, 1, 0], "response": [0, 0, 0, 1]}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 2)
        self.assertTrue(
            res.metrics[0]["key"]
            in ["seldon_metric_false_positive", "seldon_metric_false_negative"]
        )
        self.assertTrue(
            res.metrics[1]["key"]
            in ["seldon_metric_false_positive", "seldon_metric_false_negative"]
        )

        req = {"truth": [0, 0, 1, 0], "response": [0, 0, 1, 0]}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

        req = {"truth": [0.1, 0.2, 0.7, 0.1], "response": [0.1, 0.2, 0.1, 0.7]}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 2)
        self.assertTrue(
            res.metrics[0]["key"]
            in ["seldon_metric_false_positive", "seldon_metric_false_negative"]
        )
        self.assertTrue(
            res.metrics[1]["key"]
            in ["seldon_metric_false_positive", "seldon_metric_false_negative"]
        )

        req = {"truth": [0.1, 0.2, 0.7, 0.1], "response": [0.1, 0.2, 0.7, 0.1]}
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

        req = {
            "truth": [0.0006985194531162841, 0.003668039039435755, 0.9956334415074478],
            "response": [0, 0, 1],
        }
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

        req = {
            "truth": np.array(
                [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]
            ),
            "response": np.array([[0, 0, 1]]),
        }
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")

        req = {
            "truth": np.array([[0, 0, 1]]),
            "response": np.array(
                [0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]
            ),
        }
        headers = {}
        res = ad_model.process_event(req, headers)
        self.assertEqual(len(res.metrics), 1)
        self.assertEqual(res.metrics[0]["key"], "seldon_metric_true_positive")
