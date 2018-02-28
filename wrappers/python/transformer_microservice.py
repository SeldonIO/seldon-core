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

def transform_input(user_model,features,feature_names):
    if hasattr(user_model,"transform_input"):
        return user_model.transform_input(features,feature_names)
    else:
        return features

def transform_output(user_model,features,feature_names):
    if hasattr(user_model,"transform_output"):
        return user_model.transform_output(features,feature_names)
    else:
        return features

def get_feature_names(user_model,original):
    if hasattr(user_model,"feature_names"):
        return user_model.feature_names
    else:
        return original

def get_class_names(user_model,original):
    if hasattr(user_model,"class_names"):
        return user_model.class_names
    else:
        return original


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


    @app.route("/transform-input",methods=["GET","POST"])
    def TransformInput():
        request = extract_message()
        sanity_check_request(request)
        
        datadef = request.get("data")
        features = rest_datadef_to_array(datadef)

        transformed = np.array(transform_input(user_model,features,datadef.get("names")))
        # TODO: check that predictions is 2 dimensional
        new_feature_names = get_feature_names(user_model, datadef.get("names"))

        data = array_to_rest_datadef(transformed, new_feature_names, datadef)

        return jsonify({"data":data})

    @app.route("/transform-output",methods=["GET","POST"])
    def TransformOutput():
        request = extract_message()
        sanity_check_request(request)
        
        datadef = request.get("data")
        features = rest_datadef_to_array(datadef)

        transformed = np.array(transform_output(user_model,features,datadef.get("names")))
        # TODO: check that predictions is 2 dimensional
        new_class_names = get_class_names(user_model, datadef.get("names"))

        data = array_to_rest_datadef(transformed, new_class_names, datadef)

        return jsonify({"data":data})

    return app



# ----------------------------
# GRPC
# ----------------------------

class SeldonTransformerGRPC(object):
    def __init__(self,user_model):
        self.user_model = user_model

    def TransformInput(self,request,context):
        datadef = request.data
        features = grpc_datadef_to_array(datadef)

        transformed = np.array(transform_input(self.user_model,features,datadef.names))
        #TODO: check that predictions is 2 dimensional
        feature_names = get_feature_names(self.user_model, datadef.names)

        data = array_to_grpc_datadef(transformed, feature_names, request.data.WhichOneof("data_oneof"))
        return prediction_pb2.SeldonMessage(data=data)

    def TransformOutput(self,request,context):
        datadef = request.data
        features = grpc_datadef_to_array(datadef)

        transformed = np.array(transform_output(self.user_model,features,datadef.names))
        #TODO: check that predictions is 2 dimensional
        class_names = get_class_names(self.user_model, datadef.names)

        data = array_to_grpc_datadef(transformed, class_names, request.data.WhichOneof("data_oneof"))
        return prediction_pb2.SeldonMessage(data=data)
    
def get_grpc_server(user_model,debug=False):
    seldon_model = SeldonTransformerGRPC(user_model)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)

    return server
