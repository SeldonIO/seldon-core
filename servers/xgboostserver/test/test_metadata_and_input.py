import os
import tempfile
import numpy as np
import pytest
import xgboost as xgb
from XGBoostServer import XGBoostServer

@pytest.fixture
def model_uri():
    with tempfile.TemporaryDirectory() as temp_dir:
        X = np.random.rand(100, 5)
        y = np.random.randint(2, size=100)
        dtrain = xgb.DMatrix(X, label=y)
        params = {'objective': 'binary:logistic', 'eval_metric': 'error'}
        booster = xgb.train(params, dtrain, num_boost_round=10)
        model_path = os.path.join(temp_dir, "model.json")
        booster.save_model(model_path)
        yield temp_dir

def test_init_metadata(model_uri):
    metadata = {"key": "value"}
    metadata_path = os.path.join(model_uri, "metadata.yaml")
    with open(metadata_path, "w") as f:
        yaml.dump(metadata, f)

    server = XGBoostServer(model_uri)

    loaded_metadata = server.init_metadata()

    # Assert that the loaded metadata matches the original metadata
    assert loaded_metadata == metadata

def test_predict_invalid_input(model_uri):
    # Create an instance of XGBoostServer with the model URI
    server = XGBoostServer(model_uri)
    server.load()
    X_test = np.random.rand(10, 3)  # Incorrect number of features
    with pytest.raises(ValueError):
        server.predict(X_test, names=[])