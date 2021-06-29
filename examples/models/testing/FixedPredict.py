import numpy as np
import logging
from FixedBase import FixedBase

class FixedPredict(FixedBase):

    def __init__(self, iterations=1):
        super().__init__(iterations)
        
    def predict(self, X, feature_names):
        return self.work()


