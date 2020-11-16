import json

class Transformer(object):

    def transform_input(self, X, meta):
        print(X)
        return X+1
