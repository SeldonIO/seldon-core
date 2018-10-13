
class ModelWithFeedback(object):

    def __init__(self):
        print("Initialising")

    def predict(self,X,features_names):
        print("Predict called")
        return X

    def send_feedback(self,features,feature_names,reward,truth):
        print("Send feedback called")
        return []

    
