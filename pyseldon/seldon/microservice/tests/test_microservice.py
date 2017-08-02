import flask
from flask import Flask,jsonify
import unittest
from seldon.microservice.microservice import *
import sys
import logging
import numpy as np
import json

class Fake_Pipeline():

    def __init__(self):
        self._final_estimator = {}

    def predict_proba(self,X):
        y = np.array([[1.0],[2.0],[3.0]])
        return y

class Test_Microservice(unittest.TestCase):

    def setUp(self):
        m = Microservices()
        fake_pipeline = Fake_Pipeline()
        self.app = m.create_prediction_microservice(fake_pipeline,"test")
        self.app_client = self.app.test_client()

    def test_predict_nondefault(self):
        qs = {'is_default':False,'json':'{"request":{"values":"some data"}}'}
        rv = self.app_client.get('/predict',query_string=qs)
        data = json.loads(rv.data)
        self.assertEqual([3,1],data['response']['shape'])

    def test_predict_default_noshape_1dim(self):
        qs = {'is_default':True,'json':'{"request":{"values":[1.0,2.0]}}'}
        rv = self.app_client.get('/predict',query_string=qs)
        data = json.loads(rv.data)
        self.assertEqual([3,1],data['response']['shape'])

    def test_predict_default_noshape_2dim(self):
        qs = {'is_default':True,'json':'{"request":{"values":[[1.0,2.0]]}}'}
        rv = self.app_client.get('/predict',query_string=qs)
        data = json.loads(rv.data)
        self.assertEqual([3,1],data['response']['shape'])
        print data

    def test_issue(self):
        #qs = {'is_default':True,'json':'{"request":{"values":[[1.0,2.0]]}}'}
        rv = self.app_client.get('/predict?json=%7B%0A++%22meta%22%3A+%7B%0A++++%22puid%22%3A+%22rpu4v90rru9187d57b4onqua8p%22%2C%0A++++%22tags%22%3A+%7B%0A++++%7D%0A++%7D%2C%0A++%22request%22%3A+%7B%0A++++%22keys%22%3A+%5B%22a%22%5D%2C%0A++++%22shape%22%3A+%5B%5D%2C%0A++++%22values%22%3A+%5B1.0%5D%0A++%7D%0A%7D&isDefault=true')
        data = json.loads(rv.data)
        self.assertEqual([3,1],data['response']['shape'])
        print data

if __name__ == '__main__':
    logging.basicConfig(format='%(asctime)s : %(levelname)s : %(message)s', level=logging.INFO)
    unittest.main()
