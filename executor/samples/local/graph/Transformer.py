import json

class Transformer(object):

    def transform_input_raw(self, X):
        print(X)
        return json.loads(X["jsonData"])