from proto import prediction_pb2, prediction_pb2_grpc
from microservice import extract_message, sanity_check_request, rest_datadef_to_array, \
    array_to_rest_datadef, grpc_datadef_to_array, array_to_grpc_datadef, \
    SeldonMicroserviceException
import grpc
from concurrent import futures

from flask import jsonify, Flask
import numpy as np

# ---------------------------
# Interaction with user model
# ---------------------------

def predict(user_model,features,feature_names):
    return user_model.predict(features,feature_names)

def send_feedback(user_model,features,feature_names,truth,reward):
    return user_model.send_feedback(features,feature_names,truth,reward)

def get_class_names(user_model,n_targets):
    if hasattr(user_model,"class_names"):
        return user_model.class_names
    else:
        return ["t:{}".format(i) for i in range(n_targets)]


# ----------------------------
# REST
# ----------------------------

def get_rest_microservice(user_model,debug=False):

    app = Flask(__name__)

    @app.errorhandler(SeldonMicroserviceException)
    def handle_invalid_usage(error):
        response = jsonify(error.to_dict())
        print("ERROR:")
        print(error.to_dict())
        response.status_code = 400
        return response


    @app.route("/predict",methods=["GET","POST"])
    def Predict():
        request = extract_message()
        sanity_check_request(request)
        
        datadef = request.get("data")
        features = rest_datadef_to_array(datadef)

        predictions = np.array(predict(user_model,features,datadef.get("names")))
        # TODO: check that predictions is 2 dimensional
        class_names = get_class_names(user_model, predictions.shape[1])

        data = array_to_rest_datadef(predictions, class_names, datadef)

        return jsonify({"data":data})

    @app.route("/send-feedback",methods=["GET","POST"])
    def SendFeedback():
        feedback = extract_message()
        
        datadef_request = feedback.get("request").get("data")
        features = rest_datadef_to_array(datadef)
        
        truth = rest_datadef_to_array(feedback.get("truth"))
        reward = feedback.get("reward")

        send_feedback(user_model,features,datadef_request.get("names"),truth,reward)
        return jsonify({})

    return app



# ----------------------------
# GRPC
# ----------------------------

class SeldonModelGRPC(object):
    def __init__(self,user_model):
        self.user_model = user_model

    def Predict(self,request,context):
        datadef = request.data
        features = grpc_datadef_to_array(datadef)

        predictions = np.array(predict(self.user_model,features,datadef.names))
        #TODO: check that predictions is 2 dimensional
        class_names = get_class_names(self.user_model, predictions.shape[1])

        data = array_to_grpc_datadef(predictions, class_names, request.data.WhichOneof("data_oneof"))
        return prediction_pb2.SeldonMessage(data=data)

    def SendFeedback(self,feedback,context):
        datadef_request = feedback.request.data
        features = grpc_datadef_to_array(datadef_request)
        
        truth = grpc_datadef_to_array(feedback.truth)
        reward = feedback.reward

        send_feedback(self.user_model,features,datadef_request.names,truth,reward)

        return prediction_pb2.SeldonMessage()
    
def get_grpc_server(user_model,debug=False):
    seldon_model = SeldonModelGRPC(user_model)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)

    return server
