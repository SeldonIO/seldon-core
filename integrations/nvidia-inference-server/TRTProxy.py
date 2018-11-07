from inference_server.api import *
import inference_server.api.model_config_pb2 as model_config


def model_dtype_to_np(model_dtype):
    '''
    Helper function from https://github.com/NVIDIA/dl-inference-server/blob/18.08/src/clients/python/image_client.py
    '''
    if model_dtype == model_config.TYPE_BOOL:
        return np.bool
    elif model_dtype == model_config.TYPE_INT8:
        return np.int8
    elif model_dtype == model_config.TYPE_INT16:
        return np.int16
    elif model_dtype == model_config.TYPE_INT32:
        return np.int32
    elif model_dtype == model_config.TYPE_INT64:
        return np.int64
    elif model_dtype == model_config.TYPE_UINT8:
        return np.uint8
    elif model_dtype == model_config.TYPE_UINT16:
        return np.uint16
    elif model_dtype == model_config.TYPE_FP16:
        return np.float16
    elif model_dtype == model_config.TYPE_FP32:
        return np.float32
    elif model_dtype == model_config.TYPE_FP64:
        return np.float64
    return None

def parse_model(url, protocol, model_name, verbose=False):
    ctx = ServerStatusContext(url, protocol, model_name, verbose)
    server_status = ctx.get_server_status()

    if model_name not in server_status.model_status:
        raise Exception("unable to find model:"+model_name)

    status = server_status.model_status[model_name]
    config = status.config

    input = config.input[0]
    output = config.output[0]

    return (input.name, output.name, model_dtype_to_np(input.data_type), input.dims)


'''
A basic tensorflow serving proxy
'''
class TRTProxy(object):

    def __init__(self,url=None,protocol="HTTP",model_name=None,model_version=1):
        print("URL:",url)
        self.url = url
        self.protocol_id = ProtocolType.from_str(protocol)
        self.model_version = model_version
        if protocol == "GRPC":
            self.grpc = True
            channel = grpc.insecure_channel(url)
            #self.stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)
        else:
            self.grpc = False
        self.model_name = model_name
        self.input_name, self.output_name, self.dtype, self.input_dims = parse_model(url, self.protocol_id, model_name, False)
        self.ctx = InferContext(self.url, self.protocol_id,self.model_name, self.model_version, False)

        
    
    def predict(self,X,features_names):
        X = X.astype(self.dtype)
        if self.grpc:
            print("not implemented")
        else:
            if len(X.shape) == len(self.input_dims):
                X = [X]
            results = self.ctx.run(
                { self.input_name : X },
                { self.output_name : InferContext.ResultFormat.RAW },
                1)
            return results[self.output_name]
        return []
