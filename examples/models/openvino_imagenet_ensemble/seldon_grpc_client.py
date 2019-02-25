from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
import grpc
import datetime
import argparse


def grpc_request_ambassador_bindata(deploymentName,namespace,endpoint="localhost:8080",data=None):
    request = prediction_pb2.SeldonMessage(binData = data)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [('seldon',deploymentName)]
    else:
        metadata = [('seldon',deploymentName),('namespace',namespace)]
    response = stub.Predict(request=request,metadata=metadata)
    return response


def getImageBytes(path):
    with open(path, mode='rb') as file:
        fileContent = file.read()
    return fileContent


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--repeats", default=1, type=int)
    parser.add_argument('--debug', action='store_true')
    parser.add_argument('--test-input', default='input_images.txt')
    parser.add_argument('--ambassador',default='localhost:8080')

    args = parser.parse_args()

    fc = open('imagenet_classes.json')
    cnames = eval(fc.read())

    #input_images = "input_images.txt"
    input_images = args.test_input
    with open(input_images) as f:
        lines = f.readlines()
    
        i = 0
        matched = 0
        durations = []
        for j in range(args.repeats): # repeat the sequence of requests
            for line in lines:
                path, label = line.strip().split(" ")
                X = getImageBytes(path)
                start_time = datetime.datetime.now()
                response = grpc_request_ambassador_bindata("openvino-model","seldon",endpoint=args.ambassador,data=X)
                if args.debug:
                    print(response)
                end_time = datetime.datetime.now()
                duration = (end_time - start_time).total_seconds() * 1000
                durations.append(duration)
                print("Duration",duration,"ms")
                i += 1
                if response.strData == cnames[int(label)]:
                    matched += 1
        print("average duration:", sum(durations)/float(len(durations)), "ms")
        print("average accuracy:", matched/float(len(durations))*100)


if __name__ == "__main__":
    main()
