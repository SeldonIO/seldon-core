import numpy as np
from seldon_core.user_model import SeldonComponent, client_class_names
import logging

class UserObjectClassAttr(SeldonComponent):
    def __init__(self, metrics_ok=True, ret_nparray=False, ret_meta=False):
        self.class_names = ["a","b"]


class UserObjectClassMethod(SeldonComponent):
    def class_names(self):
        return ["x","y"]

def test_class_names_attr(caplog):
    caplog.set_level(logging.INFO)
    user_object = UserObjectClassAttr()
    predictions = np.array([[1,2],[3,4]])
    names = client_class_names(user_object,predictions)
    assert names == ["a","b"]
    assert "class_names attribute is deprecated. Please define a class_names method" in caplog.text

def test_class_names_method(caplog):
    caplog.set_level(logging.INFO)
    user_object = UserObjectClassMethod()
    predictions = np.array([[1, 2], [3, 4]])
    names = client_class_names(user_object, predictions)
    assert names == ["x", "y"]
    assert not "class_names attribute is deprecated. Please define a class_names method" in caplog.text