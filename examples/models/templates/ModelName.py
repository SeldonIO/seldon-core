from <your_loading_library> import <your_loading_function>

class ModelName(object):

    def __init__(self):
        self.model = <your_loading_function>(<your_model_file>)

    def predict(self,X,features_names):
        return self.model.predict(X)
