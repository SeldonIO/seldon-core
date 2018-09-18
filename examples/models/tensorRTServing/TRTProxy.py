from inference_server.api import *
import inference_server.api.model_config_pb2 as model_config

MEANS=np.array([255.0,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,254,254,254,253,252,252,251,251,252,252,253,254,254,255,255,255,255,255,255,255,255,255,255,255,255,255,254,254,253,251,249,248,245,243,242,242,243,246,248,251,253,254,255,255,255,255,255,255,255,255,255,255,255,254,253,250,247,242,235,228,220,213,210,211,216,224,232,240,246,251,253,254,255,255,255,255,255,255,255,255,254,251,248,242,234,223,211,196,181,170,164,166,175,189,205,221,233,243,248,252,254,255,255,255,255,255,255,254,252,248,241,231,217,202,184,166,149,136,131,134,143,159,180,201,220,234,243,249,253,255,255,255,255,255,254,253,249,243,233,219,201,181,161,143,130,122,120,122,129,141,161,185,208,227,240,248,252,254,255,255,255,255,254,251,246,238,226,208,187,164,146,135,131,132,133,132,133,139,154,178,202,223,239,248,252,255,255,255,255,254,253,251,245,236,221,200,177,156,144,144,150,156,156,151,144,144,156,178,202,224,240,249,253,255,255,255,255,254,253,251,245,235,218,195,172,155,152,161,172,176,170,161,150,149,161,183,207,227,242,250,254,255,255,255,255,255,254,251,246,234,215,191,168,156,160,173,182,179,169,157,147,149,166,190,213,230,243,251,254,255,255,255,255,255,254,252,246,233,212,186,165,157,164,175,176,165,153,142,137,147,170,196,217,231,242,251,255,255,255,255,255,255,254,252,245,230,207,182,163,158,164,168,158,143,131,125,128,146,174,200,218,231,241,250,254,255,255,255,255,255,255,252,243,227,205,181,164,159,161,157,139,124,115,118,127,148,176,199,216,230,240,249,254,255,255,255,255,255,254,251,241,224,204,184,169,163,160,150,132,119,116,123,133,153,177,197,214,228,240,249,254,255,255,255,255,255,254,251,239,222,205,189,177,171,166,154,139,129,128,134,144,159,177,195,213,228,241,249,254,255,255,255,255,255,254,249,237,222,207,195,186,180,175,166,153,143,140,142,150,162,178,195,214,230,242,250,254,255,255,255,255,255,253,247,235,220,207,197,189,183,179,172,160,148,142,143,150,161,178,198,217,233,244,250,254,255,255,255,255,255,253,246,233,218,204,192,184,177,172,165,153,142,137,139,148,163,183,204,222,236,246,251,254,255,255,255,255,255,253,247,234,218,201,186,174,165,157,148,137,130,129,137,151,171,194,214,230,242,248,252,254,255,255,255,255,255,253,249,238,222,203,184,168,154,143,132,124,123,130,145,165,188,209,227,239,247,251,253,255,255,255,255,255,255,254,251,244,232,214,194,174,156,142,132,130,134,148,167,189,210,226,238,246,250,253,254,255,255,255,255,255,255,255,253,250,243,231,215,196,178,163,155,156,164,179,197,215,230,240,247,251,253,254,255,255,255,255,255,255,255,255,254,253,251,246,238,228,217,208,203,204,210,218,228,236,243,248,251,253,254,255,255,255,255,255,255,255,255,255,255,255,254,252,249,245,241,238,237,237,239,242,245,247,250,252,253,254,255,255,255,255,255,255,255,255,255,255,255,255,254,254,253,252,250,249,248,249,249,250,252,253,253,254,254,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,254,254,254,254,255,255,255,255,255,255,255,255,255,255,255,255])


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

def parse_model(url, protocol, model_name, batch_size, verbose=False):
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
        self.input_name, self.output_name, self.dtype, self.input_dims = parse_model(url, self.protocol_id, model_name,1 , False)

        
    def preProcessMNIST(self,X):
        '''
        Example single mnist transform. Move to separate component.
        '''
        X = X * 255
        X = 255 - X
        X = (X.reshape(784) - MEANS).reshape(28,28,1)
        X = X.astype(self.dtype)
        X = np.transpose(X, (2, 0, 1))
        return X

    def postProcessMNIST(self,X):
        '''
        Example single mnist post transform. Move to separate component.
        '''
        X = np.array(X)
        return X.reshape(1,10)
    
    def predict(self,X,features_names):
        if self.grpc:
            print("not implemented")
        else:
            X = self.preProcessMNIST(X)
            if len(X.shape) == len(self.input_dims):
                X = [X]
            ctx = InferContext(self.url, self.protocol_id,self.model_name, self.model_version, False)
            results = ctx.run(
                { self.input_name : X },
                { self.output_name : InferContext.ResultFormat.RAW },
                1)
            return self.postProcessMNIST(results[self.output_name])
        return []
