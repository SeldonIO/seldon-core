from proto import prediction_pb2, prediction_pb2_grpc
from microservice import extract_message, sanity_check_request, rest_datadef_to_array, \
    array_to_rest_datadef, grpc_datadef_to_array, array_to_grpc_datadef, \
    SeldonMicroserviceException
import grpc
from concurrent import futures

from flask import jsonify, Flask
from flask_cors import CORS
import numpy as np
import logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

from tornado.tcpserver import TCPServer
from tornado.iostream import StreamClosedError
from tornado import gen
import tornado.ioloop
from seldon_flatbuffers import SeldonRPCToNumpyArray,NumpyArrayToSeldonRPC,CreateErrorMsg
import struct
import traceback
import os

PRED_UNIT_ID = os.environ.get("PREDICTIVE_UNIT_ID")

# ---------------------------
# Interaction with user model
# ---------------------------

def predict(user_model,features,feature_names):
    return user_model.predict(features,feature_names)

def send_feedback(user_model,features,feature_names,reward,truth):
    if hasattr(user_model,"send_feedback"):
        user_model.send_feedback(features,feature_names,reward,truth)

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
    CORS(app)
    
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
        if len(predictions.shape)>1:
            class_names = get_class_names(user_model, predictions.shape[1])
        else:
            class_names = []
            
        data = array_to_rest_datadef(predictions, class_names, datadef)

        return jsonify({"data":data})

    @app.route("/send-feedback",methods=["GET","POST"])
    def SendFeedback():
        feedback = extract_message()

        datadef_request = feedback.get("request",{}).get("data",{})
        features = rest_datadef_to_array(datadef_request)
        
        truth = rest_datadef_to_array(feedback.get("truth",{}))
        reward = feedback.get("reward")

        send_feedback(user_model,features,datadef_request.get("names"),reward,truth)        
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
        if len(predictions.shape)>1:
            class_names = get_class_names(self.user_model, predictions.shape[1])
        else:
            class_names = []

        data = array_to_grpc_datadef(predictions, class_names, request.data.WhichOneof("data_oneof"))
        return prediction_pb2.SeldonMessage(data=data)

    def SendFeedback(self,feedback,context):
        datadef_request = feedback.request.data
        features = grpc_datadef_to_array(datadef_request)
        
        truth = grpc_datadef_to_array(feedback.truth)
        reward = feedback.reward

        send_feedback(self.user_model,features,datadef_request.names,truth,reward)

        return prediction_pb2.SeldonMessage()

ANNOTATION_GRPC_MAX_MSG_SIZE = 'seldon.io/grpc-max-message-size'
    
def get_grpc_server(user_model,debug=False,annotations={}):
    seldon_model = SeldonModelGRPC(user_model)
    options = []
    if ANNOTATION_GRPC_MAX_MSG_SIZE in annotations:
        max_msg = int(annotations[ANNOTATION_GRPC_MAX_MSG_SIZE])
        logger.info("Setting grpc max message and receive length to %d",max_msg)
        options.append(('grpc.max_message_length', max_msg ))
        options.append(('grpc.max_receive_message_length', max_msg))

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10),options=options)
    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)

    return server


# ----------------------------
# Flatbuffers (experimental)
# ----------------------------

class SeldonFlatbuffersServer(TCPServer):
    def __init__(self,user_model):
        super(SeldonFlatbuffersServer, self).__init__()
        self.user_model = user_model

    @gen.coroutine
    def handle_stream(self, stream, address):
        while True:
            try:
                data = yield stream.read_bytes(4)
                obj = struct.unpack('<i',data)
                len_msg = obj[0]
                data = yield stream.read_bytes(len_msg)
                try:
                    features,names = SeldonRPCToNumpyArray(data)
                    predictions = np.array(predict(self.user_model,features,names))
                    if len(predictions.shape)>1:
                        class_names = get_class_names(self.user_model, predictions.shape[1])
                    else:
                        class_names = []
                    outData = NumpyArrayToSeldonRPC(predictions,class_names)
                    yield stream.write(outData)
                except StreamClosedError:
                    print("Stream closed during processing:",address)
                    break
                except Exception:
                    tb = traceback.format_exc()
                    print("Caught exception during processing:",address,tb)
                    outData = CreateErrorMsg(tb)
                    yield stream.write(outData)
                    stream.close()
                    break;
            except StreamClosedError:
                print("Stream closed during data inputstream read:",address)
                break
        
def run_flatbuffers_server(user_model,port,debug=False):
    server = SeldonFlatbuffersServer(user_model)
    server.listen(port)
    print("Tornando Server listening on port",port)
    tornado.ioloop.IOLoop.current().start()
