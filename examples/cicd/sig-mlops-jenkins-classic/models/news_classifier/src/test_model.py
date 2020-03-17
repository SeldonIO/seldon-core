import numpy as np
from unittest import mock
import joblib
import os

EXPECTED_RESPONSE = np.array([3, 3])


def test_model(*args, **kwargs):
    data = ["text 1", "text 2"]

    m = joblib.load("model.joblib")
    result = m.predict(data)
    assert all(result == EXPECTED_RESPONSE)
