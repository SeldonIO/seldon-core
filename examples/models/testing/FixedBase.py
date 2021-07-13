import numpy as np
import logging

class FixedBase(object):

    def __init__(self, iterations=1000000):
        assert type(iterations) == int, "intValue parameter must be an integer"
        self.iterations = iterations
        logging.info("iterations set to %d",iterations)
        
    def work(self):
        work = 0
        for i in range(0,self.iterations):
            work = work + 1
        return [work]

    def health_status(self):
        return {"status":"ok"}

