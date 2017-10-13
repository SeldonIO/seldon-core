from seldon_model import SeldonModel

class MyModel(SeldonModel):
    def __init__(self):
        print "Loading weights, etc"

    def predict(self,features):
        return sum(features)
