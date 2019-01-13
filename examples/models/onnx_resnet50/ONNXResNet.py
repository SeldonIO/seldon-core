from keras.applications.imagenet_utils import decode_predictions
from keras.applications.resnet50 import preprocess_input
from keras.preprocessing import image
import ngraph as ng
import numpy as np
from ngraph_onnx.onnx_importer.importer import import_onnx_file

class ONNXResNet(object):

    def __init__(self):
        print("Loading model")
        # Import the ONNX file
        models = import_onnx_file('resnet50/model.onnx')
        # Create an nGraph runtime environment
        runtime = ng.runtime(backend_name='CPU')
        # Select the first model and compile it to a callable function
        model = models[0]
        self.resnet = runtime.computation(model['output'], *model['inputs'])
        print("Model loaded")

        #Do a test run to warm up and check all is ok
        print("Running test on img of Zebra as warmup")
        img = image.load_img('zebra.jpg', target_size=(224, 224))
        img = image.img_to_array(img)
        x = np.expand_dims(img.copy(), axis=0)
        x = preprocess_input(x,mode='torch')
        x = x.transpose(0,3,1,2)
        preds = self.resnet(x)
        print(decode_predictions(preds[0], top=5))
        
    def predict(self,X,features_names):
        print(X.shape)
        X = preprocess_input(X,mode='torch')
        preds = self.resnet(X)
        print(decode_predictions(preds[0], top=5))
        return preds
