from proto import prediction_pb2
import numpy as np

def get_grpc_features(datadef):
    if len(datadef.tensor.values)>0:
        return np.array(datadef.tensor.values).reshape(datadef.tensor.shape),datadef.names
    else:
        return np.array(datadef.ndarray),datadef.names

def gen_grpc_response(predictions,target_names,request):
    if len(request.request.tensor.values)>0:
        # We will return a tensor in the response

        response = prediction_pb2.ResponseDef(
            response=prediction_pb2.DefaultDataDef(
                names = target_names,
                tensor = prediction_pb2.Tensor(
                    shape = predictions.shape,
                    values = predictions.ravel().tolist()
                )
            )
        )
    else:
        # We will return a ndarray in the response
        
        response = prediction_pb2.ResponseDef(
            response=prediction_pb2.DefaultDataDef(
                names = target_names,
                ndarray = predictions.tolist()
            )
        )

    return response

def get_rest_features(data):
    if data.get("tensor") is not None:
        shape = data["tensor"].get("shape")
        values = np.array(data["tensor"].get("values")).reshape(shape)
    else:
        values = np.array(data.get("ndarray"))
    return values, data.get("names")
    

def gen_rest_response(preds, names, tensor=True):
    response = {'names':names}
    if tensor:
        response['tensor'] = {
            'shape':preds.shape,
            'values':preds.ravel().tolist()
        }
    else:
        response['ndarray'] = preds.tolist()
    ret = {'response':response}
    return ret
    
class SeldonModel(object):
    def __init__(self,UserModelClass,parameters):
        self.user_model = UserModelClass(**parameters)

    def get_class_names(self,n_targets):
        if hasattr(self.user_model,"class_names"):
            return self.user_model.class_names
        else:
            return ["t:{}".format(i) for i in range(n_targets)]
        
    def predict(self,features,feature_names):
        return self.user_model.predict(features,feature_names)

    def feedback(self,feedback):
        return self.user_model.feedback(feedback)
                 
    def predict_rest(self,data,is_default):
        if is_default:
            features, feature_names = get_rest_features(data)
        else:
            features = data
            features_names = []
        predictions = np.array(self.predict(features, feature_names))
        n_targets = predictions.shape[1]
        if is_default:
            return gen_rest_response(
                predictions,
                self.get_class_names(n_targets),
                tensor=data.get("tensor") is not None)
        else:
            return gen_rest_response(predictions, self.get_class_names(n_targets))

    def Predict(self,request,context):
        features, feature_names = get_grpc_features(request.request)
        predictions = np.array(self.predict(features, feature_names))
        n_targets = predictions.shape[1]
        return gen_grpc_response(predictions, self.get_class_names(n_targets), request)

    def Feedback(self,feedback,context):
        self.feedback(feedback)
        return prediction_pb2.ResponseDef()
