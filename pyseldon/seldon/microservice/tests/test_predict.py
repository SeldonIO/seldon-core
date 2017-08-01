import flask
from flask import Flask,jsonify
import unittest
from seldon.microservice.predict import *
import sys
import logging

app = Flask(__name__)

class Test_Predict(unittest.TestCase):

    def test_ndarray_12dim(self):
        data = {"ndarray":[[1.0,2.0]]}
        df = default_data_to_dataframe(data)
        self.assertEqual(df.shape,(1,2))

    def test_ndarray_1dim(self):
        data = {"ndarray":[1.0,2.0]}
        self.assertRaises(DataContractException,default_data_to_dataframe,data)

    def test_ndarray_22dim(self):
        data = {"ndarray":[[1.0,2.0],[3.0,4.0]]}
        df = default_data_to_dataframe(data)
        self.assertEqual(df.shape,(2,2))


    def test_no_shape_1dim(self):
        data = {"values":[1.0,2.0]}
        df = default_data_to_dataframe(data)
        self.assertEqual(df.shape,(1,2))

    def test_no_shape_2dim(self):
        data = {"values":[[1.0,2.0],[3.0,4.0]]}
        df = default_data_to_dataframe(data)
        self.assertEqual(df.shape,(2,2))

    def test_shape_1dim(self):
        data = {"values":[1.0,2.0],"shape":[1,2]}
        df = default_data_to_dataframe(data)
        self.assertEqual(df.shape,(1,2))

    def test_shape_2dim(self):
        data = {"values":[1.0,2.0,3.0,4.0],"shape":[2,2]}
        df = default_data_to_dataframe(data)
        self.assertEqual(df.shape,(2,2))
        
    @app.route('/kw')
    def test_json(self):
        preds = np.array([[1.0,2.0],[3.0,4.0]])
        names = ["a","b"]
        ret = create_response(names,preds)
        app.config['TESTING'] = True
        c = app.test_client()
        with app.test_request_context('/?name=Peter'):
            json_ret = flask.jsonify(ret)

if __name__ == '__main__':
    logging.basicConfig(format='%(asctime)s : %(levelname)s : %(message)s', level=logging.INFO)
    unittest.main()
