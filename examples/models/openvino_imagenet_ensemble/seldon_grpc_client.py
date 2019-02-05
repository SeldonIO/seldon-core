from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
import grpc
import datetime


API_AMBASSADOR="localhost:8080"


def grpc_request_ambassador_bindata(deploymentName,namespace,endpoint="localhost:8004",data=None):
    request = prediction_pb2.SeldonMessage(binData = data)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [('seldon',deploymentName)]
    else:
        metadata = [('seldon',deploymentName),('namespace',namespace)]
    response = stub.Predict(request=request,metadata=metadata)
    return response


def getImage(path):
    img = image.load_img(path, target_size=(227, 227))
    x = image.img_to_array(img)
    x = np.expand_dims(x, axis=0)
    x = preprocess_input(x)
    return x

def getImageBytes(path):
    with open(path, mode='rb') as file:
        fileContent = file.read()
    return fileContent


fc = open('imagenet_classes.json')
cnames = eval(fc.read())
print(type(cnames))

input_images = "input_images.txt"
with open(input_images) as f:
    lines = f.readlines()

i = 0
matched = 0
durations = []
for j in range(20): # repeat the sequence of requests
    for line in lines:
        path, label = line.strip().split(" ")
        X = getImageBytes(path)
        start_time = datetime.datetime.now()
        response = grpc_request_ambassador_bindata("openvino-model","default",API_AMBASSADOR,data=X)
        end_time = datetime.datetime.now()
        duration = (end_time - start_time).total_seconds() * 1000
        durations.append(duration)
        print("Duration",duration)
        i += 1
        if response.strData == cnames[int(label)]:
            matched += 1
print("average duration:",sum(durations)/float(len(durations)))
print("average accuracy:",matched/float(len(durations))*100)