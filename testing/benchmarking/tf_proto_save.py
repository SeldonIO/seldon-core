import argparse
import base64
import json

from tensorflow.contrib.util import make_tensor_proto
from tensorflow_serving.apis import predict_pb2


def serialize(model, signature_name, input_path, output_path):
    # If hostname not set, we assume the host is a valid knative dns.

    with open(input_path) as json_file:
        data = json.load(json_file)
    image = data["instances"][0]["image_bytes"]["b64"]
    key = data["instances"][0]["key"]

    # Call classification model to make prediction
    request = predict_pb2.PredictRequest()
    request.model_spec.name = model
    request.model_spec.signature_name = signature_name
    image = base64.b64decode(image)
    request.inputs["image_bytes"].CopyFrom(make_tensor_proto(image, shape=[1]))
    request.inputs["key"].CopyFrom(make_tensor_proto(key, shape=[1]))

    bin_file = open(output_path, "wb")
    bin_file.write(request.SerializeToString())
    bin_file.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", help="TensorFlow Model Name", type=str)
    parser.add_argument(
        "--signature_name",
        help="Signature name of saved TensorFlow model",
        default="serving_default",
        type=str,
    )
    parser.add_argument(
        "--input_path",
        help="Prediction data input path",
        default="./input.json",
        type=str,
    )
    parser.add_argument(
        "--output_path",
        help="Prediction data output path",
        default="./payload.bin",
        type=str,
    )

    args = parser.parse_args()
    serialize(args.model, args.signature_name, args.input_path, args.output_path)
