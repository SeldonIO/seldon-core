import unittest

import os
import default_logger
import numpy as np
import log_helper
import json

cors_enabled = os.getenv("CORS_ENABLED", "False")

testapp = default_logger.app.test_client()

# for more local testing see README
class TestRequestLogger(unittest.TestCase):
    def test_enriched_elements_request(self):
        """
        Test that an elements array can be built using metadata to enrich a request
        """

        # mixture of one_hot, categorical and float
        names = [
            "dummy_one_hot_1",
            "dummy_one_hot_2",
            "dummy_categorical",
            "dummy_float",
        ]
        X = np.array([[0.0, 1.0, 0.0, 2.54]])
        results = None
        metadata_schema = {
            "requests": [
                {
                    "name": "dummy_one_hot",
                    "type": "ONE_HOT",
                    "data_type": "INT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [
                        {"name": "dummy_one_hot_1", "data_type": "FLOAT"},
                        {"name": "dummy_one_hot_2", "data_type": "FLOAT"},
                    ],
                    "shape": [],
                },
                {
                    "name": "dummy_categorical",
                    "type": "CATEGORICAL",
                    "data_type": "INT",
                    "n_categories": "2",
                    "category_map": {"0": "dummy_cat_0", "1": "dummy_cat_1"},
                    "schema": [],
                    "shape": [],
                },
                {
                    "name": "dummy_float",
                    "type": "REAL",
                    "data_type": "FLOAT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [],
                    "shape": [],
                },
            ],
            "responses": [
                {
                    "name": "dummy_proba",
                    "type": "PROBA",
                    "data_type": "FLOAT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [
                        {"name": "dummy_proba_0", "data_type": "FLOAT"},
                        {"name": "dummy_proba_1", "data_type": "FLOAT"},
                    ],
                    "shape": [],
                },
                {
                    "name": "dummy_float",
                    "type": "REAL",
                    "data_type": "FLOAT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [],
                    "shape": [],
                },
            ],
        }
        message_type = "request"

        # one_hot columns should be merged under a top-level, categorical replaced with label and float as-is
        expected_results = [
            {
                "dummy_one_hot": {"dummy_one_hot_1": 0.0, "dummy_one_hot_2": 1.0},
                "dummy_categorical": "dummy_cat_0",
                "dummy_float": 2.54,
            }
        ]
        actual_results = default_logger.createElementsWithMetadata(
            X, names, results, metadata_schema, message_type
        )
        self.assertEqual(expected_results, actual_results)

    def test_enriched_elements_response(self):
        """
        Test that an elements array can be built using metadata to enrich a response
        """

        # mixture of proba and float
        names = ["dummy_proba_0", "dummy_proba_1", "dummy_float"]
        X = np.array([[0.85388188, 0.14611812, 3.65]])
        results = None
        metadata_schema = {
            "requests": [
                {
                    "name": "dummy_one_hot",
                    "type": "ONE_HOT",
                    "data_type": "INT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [
                        {"name": "dummy_one_hot_1", "data_type": "FLOAT"},
                        {"name": "dummy_one_hot_2", "data_type": "FLOAT"},
                    ],
                    "shape": [],
                },
                {
                    "name": "dummy_categorical",
                    "type": "CATEGORICAL",
                    "data_type": "INT",
                    "n_categories": "2",
                    "category_map": {"0": "dummy_cat_0", "1": "dummy_cat_1"},
                    "schema": [],
                    "shape": [],
                },
                {
                    "name": "dummy_float",
                    "type": "REAL",
                    "data_type": "FLOAT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [],
                    "shape": [],
                },
            ],
            "responses": [
                {
                    "name": "dummy_proba",
                    "type": "PROBA",
                    "data_type": "FLOAT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [
                        {"name": "dummy_proba_0", "data_type": "FLOAT"},
                        {"name": "dummy_proba_1", "data_type": "FLOAT"},
                    ],
                    "shape": [],
                },
                {
                    "name": "dummy_float",
                    "type": "REAL",
                    "data_type": "FLOAT",
                    "n_categories": "0",
                    "category_map": {},
                    "schema": [],
                    "shape": [],
                },
            ],
        }
        message_type = "response"

        # proba columns should be merged under a top-level and float as-is
        expected_results = [
            {
                "dummy_proba": {
                    "dummy_proba_0": 0.85388188,
                    "dummy_proba_1": 0.14611812,
                },
                "dummy_float": 3.65,
            }
        ]
        actual_results = default_logger.createElementsWithMetadata(
            X, names, results, metadata_schema, message_type
        )
        self.assertEqual(expected_results, actual_results)

    def test_not_enriched_elements_request(self):
        """
        Test that an elements array can be built even without metadata, provided names given
        """

        # mixture of one_hot, categorical and float
        names = [
            "Age",
            "Workclass",
            "Education",
            "Marital Status",
            "Occupation",
            "Relationship",
            "Race",
            "Sex",
            "Capital Gain",
            "Capital Loss",
            "Hours per week",
            "Country",
        ]
        X = np.array([[53.0, 4.0, 0.0, 2.0, 8.0, 4.0, 2.0, 0.0, 0.0, 0.0, 60.0, 9.0]])
        results = None

        # values should be matched to names
        expected_results = [
            {
                "Age": 53.0,
                "Workclass": 4.0,
                "Education": 0.0,
                "Marital Status": 2.0,
                "Occupation": 8.0,
                "Relationship": 4.0,
                "Race": 2.0,
                "Sex": 0.0,
                "Capital Gain": 0.0,
                "Capital Loss": 0.0,
                "Hours per week": 60.0,
                "Country": 9.0,
            }
        ]
        actual_results = default_logger.createElementsNoMetadata(X, names, results)
        self.assertEqual(expected_results, actual_results)

    def test_no_ce_requied_headers(self):
        response = testapp.post(
            "/", data=dict(data=dict(ndarray=[[1]])), follow_redirects=True
        )
        self.assertTrue(response.status_code == 400)

    def test_seldon_request(self):
        response = testapp.post(
            "/",
            data='{"data":{"ndarray":[[1]]}}',
            follow_redirects=True,
            headers=[
                (log_helper.TYPE_HEADER_NAME, "io.seldon.serving.inference.request"),
                (log_helper.REQUEST_ID_HEADER_NAME, "1"),
                (log_helper.MODELID_HEADER_NAME, self.test_seldon_request.__name__),
            ],
        )
        self.assertTrue(response.status_code == 200)
        contents_added = response.json
        for item in contents_added:
            document = default_logger.es.get(item["_index"], item["_id"])
            data = document["_source"]
            self.assertTrue(data["request"]["dataType"] == "number")

    def test_seldon_jsondata_request(self):
        response = testapp.post(
            "/",
            data=json.dumps({"jsonData": [1, 2, 3, 5]}),
            follow_redirects=True,
            headers=[
                (log_helper.TYPE_HEADER_NAME, "io.seldon.serving.inference.request"),
                (log_helper.REQUEST_ID_HEADER_NAME, "1"),
                (
                    log_helper.MODELID_HEADER_NAME,
                    self.test_seldon_jsondata_request.__name__,
                ),
            ],
        )
        self.assertTrue(response.status_code == 200)
        contents_added = response.json
        self.assertTrue(len(contents_added) == 1)

    def test_seldon_requests_ok(self):
        payloads = [
            '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}}',
            '{"data":{"names":["a","b"],"ndarray":[[1,2],[3,4]]}}',
            '{"data":{"names":["a"],"ndarray":["test1","test2"]}}',
            '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}}',
        ]
        for idx, payload in enumerate(payloads):
            response = testapp.post(
                "/",
                data=json.dumps({"jsonData": [1, 2, 3, 5]}),
                follow_redirects=True,
                headers=[
                    (
                        log_helper.TYPE_HEADER_NAME,
                        "io.seldon.serving.inference.request",
                    ),
                    (log_helper.REQUEST_ID_HEADER_NAME, "1"),
                    (
                        log_helper.MODELID_HEADER_NAME,
                        self.test_seldon_requests_ok.__name__ + str(idx),
                    ),
                ],
            )
            self.assertTrue(response.status_code == 200)

    def test_seldon_responses_ok(self):
        payloads = [
            '{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}',
            '{"data":{"names":["c"],"ndarray":[[7],[8]]}}',
            '{"data":{"names":["c"],"ndarray":[[7],[8]]}}',
            '{"data":{"names":["t0","t1"],"ndarray":[[0.5,0.5]]}}',
        ]
        for idx, payload in enumerate(payloads):
            response = testapp.post(
                "/",
                data=json.dumps({"jsonData": [1, 2, 3, 5]}),
                follow_redirects=True,
                headers=[
                    (
                        log_helper.TYPE_HEADER_NAME,
                        "io.seldon.serving.inference.response",
                    ),
                    (log_helper.REQUEST_ID_HEADER_NAME, "1"),
                    (
                        log_helper.MODELID_HEADER_NAME,
                        self.test_seldon_requests_ok.__name__ + str(idx),
                    ),
                ],
            )
            self.assertTrue(response.status_code == 200)

    def test_cors_preflight_ok(self):
        if cors_enabled.lower() in ("true", "1"):
            res = testapp.options(
                "/",
                headers=[
                    ("Access-Control-Request-Method", "POST"),
                    ("Access-Control-Request-Headers", "content-type"),
                    ("Origin", "http://example.com"),
                ],
            )
            origin = res.request.headers.get("Origin")

            self.assertTrue(
                res.headers.get("Access-Control-Allow-Origin") in ("*", origin)
            )


if __name__ == "__main__":
    unittest.main()
