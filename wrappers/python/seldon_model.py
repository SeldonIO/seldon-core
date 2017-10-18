from proto import prediction_pb2
import numpy as np

def get_features(datadef):
    if len(datadef.tensor.values)>0:
        return np.array(datadef.tensor.values).reshape(datadef.tensor.shape)
    else:
        return np.array(datadef.ndarray)

def gen_response(predictions,target_names,request):
    if len(request.request.tensor.values)>0:
        # We will return a tensor in the response

        response = prediction_pb2.PredictionResponseDef(
            response=prediction_pb2.DefaultDataDef(
                features = target_names,
                tensor = prediction_pb2.Tensor(
                    shape = predictions.shape,
                    values = predictions.ravel().tolist()
                )
            )
        )
    else:
        # We will return a ndarray in the response
        
        response = prediction_pb2.PredictionResponseDef(
            response=prediction_pb2.DefaultDataDef(
                features = target_names,
                ndarray = predictions.tolist()
            )
        )

    return response
    
class SeldonModel(object):
    def __init__(self):
        pass

    def predict(self):
        pass

    def Predict(self,request,context):
        features = get_features(request.request)
        predictions = self.predict(features)
        return gen_response(np.array(predictions), self.class_names, request)

        
