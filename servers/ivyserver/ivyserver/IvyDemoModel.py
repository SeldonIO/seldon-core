import ivy

#seldon wrapper class
class IvyDemoModel():
    def __init__(self):
        self._model = MLP()
    
    def predict(self, X):
        x_in = ivy.array(X)
        y_out = self._model(x_in)
        return y_out
