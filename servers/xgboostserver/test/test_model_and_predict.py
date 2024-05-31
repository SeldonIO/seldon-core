import os
import tempfile
import numpy as np
from unittest import mock
import xgboost as xgb
from XGBoostServer import XGBoostServer

def test_load_json_model():

    with tempfile.TemporaryDirectory() as temp_dir:
        X = np.random.rand(100, 5)
        y = np.random.randint(2, size=100)
        dtrain = xgb.DMatrix(X, label=y)
        params = {'objective': 'binary:logistic', 'eval_metric': 'error'}
        booster = xgb.train(params, dtrain, num_boost_round=10)
        model_path = os.path.join(temp_dir, "model.json")
        booster.save_model(model_path)

        server = XGBoostServer(temp_dir)
        server.load()

        assert server.ready
        assert isinstance(server._booster, xgb.Booster)

def test_predict():
    # Create a temporary directory for the model file
    with tempfile.TemporaryDirectory() as temp_dir:
        # Train a dummy XGBoost model and save it in .json format
        X = np.random.rand(100, 5)
        y = np.random.randint(2, size=100)
        dtrain = xgb.DMatrix(X, label=y)
        params = {'objective': 'binary:logistic', 'eval_metric': 'error'}
        booster = xgb.train(params, dtrain, num_boost_round=10)
        model_path = os.path.join(temp_dir, "model.json")
        booster.save_model(model_path)

        # Create an instance of XGBoostServer with the model URI
        server = XGBoostServer(temp_dir)

        server.load()

        # Prepare test data
        X_test = np.random.rand(10, 5)

        with mock.patch("seldon_core.Storage.download", return_value=temp_dir):
            predictions = server.predict(X_test, names=[])

        # Assert the expected shape and type of predictions
        assert isinstance(predictions, np.ndarray)
        assert predictions.shape == (10,)