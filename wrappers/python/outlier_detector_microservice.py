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

def score(user_model,features,feature_names):
    # Returns a single float that corresponds to the outlier score
    return user_model.score(features,feature_names)
    
# ----------------------------
# REST
# ----------------------------

def get_rest_microservice(user_model):

    app = Flask(__name__)

    @app.errorhandler(SeldonMicroserviceException)
    def handle_invalid_usage(error):
        response = jsonify(error.to_dict())
        print "ERROR:"
        print error.to_dict()
        response.status_code = 400
        return response


    @app.route("/transform-input",methods=["GET","POST"])
    def TransformInput():
        request = extract_message()
        sanity_check_request(request)
        
        datadef = request.get("data")
        features = rest_datadef_to_array(datadef)

        outlier_score = score(user_model,features,datadef.get("names"))
        # TODO: check that predictions is 2 dimensional

        request["meta"]["tags"]["outlierScore"] = outlierScore

        return jsonify(request)



# ----------------------------
# GRPC
# ----------------------------

class SeldonTransformerGRPC(object):
    def __init__(self,user_model):
        self.user_model = user_model

    def TransformInput(self,request,context):
        datadef = request.data
        features = grpc_datadef_to_array(datadef)

        outlier_score = score(self.user_model,features,datadef.names)

        request.meta.tags["outlierScore"] = outlier_score

        return request
    
def get_grpc_server(user_model):
    seldon_model = SeldonTransformerGRPC(user_model)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)

    return server
