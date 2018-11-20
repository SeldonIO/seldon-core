import argparse
import requests
from requests.auth import HTTPBasicAuth
import numpy as np
import json
import requests
import urllib
from google.protobuf import json_format
from google.protobuf.struct_pb2 import ListValue
import grpc
import sys

from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc


def array_to_list_value(array, lv=None):
    if lv is None:
        lv = ListValue()
    if len(array.shape) == 1:
        lv.extend(array)
    else:
        for sub_array in array:
            sub_lv = lv.add_list()
            array_to_list_value(sub_array, sub_lv)
    return lv


def gen_continuous(range, n):
    if range[0] == "inf" and range[1] == "inf":
        return np.random.normal(size=n)
    if range[0] == "inf":
        return range[1] - np.random.lognormal(size=n)
    if range[1] == "inf":
        return range[0] + np.random.lognormal(size=n)
    return np.random.uniform(range[0], range[1], size=n)


def reconciliate_cont_type(feature, dtype):
    if dtype == "FLOAT":
        return feature
    if dtype == "INT":
        return (feature + 0.5).astype(int).astype(float)


def gen_categorical(values, n):
    vals = np.random.randint(len(values), size=n)
    return np.array(values)[vals].astype(float)


def generate_batch(contract, n, field):
    feature_batches = []
    for feature_def in contract[field]:
        if feature_def["ftype"] == "continuous":
            if "range" in feature_def:
                range = feature_def["range"]
            else:
                range = ["inf", "inf"]
            if "shape" in feature_def:
                shape = [n] + feature_def["shape"]
            else:
                shape = [n, 1]
            batch = gen_continuous(range, shape)
            batch = reconciliate_cont_type(batch, feature_def["dtype"])
        elif feature_def["ftype"] == "categorical":
            batch = gen_categorical(feature_def["values"], [n, 1])
        feature_batches.append(batch)
    return np.concatenate(feature_batches, axis=1)


def gen_REST_request(batch, features, tensor=True):
    if tensor:
        datadef = {
            "names": features,
            "tensor": {
                "shape": batch.shape,
                "values": batch.ravel().tolist()
            }
        }
    else:
        datadef = {
            "names": features,
            "ndarray": batch.tolist()
        }

    request = {
        "meta": {},
        "data": datadef
    }

    return request


def gen_GRPC_request(batch, features, tensor=True):
    if tensor:
        datadef = prediction_pb2.DefaultData(
            names=features,
            tensor=prediction_pb2.Tensor(
                shape=batch.shape,
                values=batch.ravel().tolist()
            )
        )
    else:
        datadef = prediction_pb2.DefaultData(
            names=features,
            ndarray=array_to_list_value(batch)
        )
    request = prediction_pb2.SeldonMessage(
        data=datadef
    )
    return request


def unfold_contract(contract):
    unfolded_contract = {}
    unfolded_contract["targets"] = []
    unfolded_contract["features"] = []

    for feature in contract["features"]:
        if feature.get("repeat") is not None:
            for i in range(feature.get("repeat")):
                new_feature = {}
                new_feature.update(feature)
                new_feature["name"] = feature["name"] + str(i + 1)
                del new_feature["repeat"]
                unfolded_contract["features"].append(new_feature)
        else:
            unfolded_contract["features"].append(feature)

    for target in contract["targets"]:
        if target.get("repeat") is not None:
            for i in range(target.get("repeat")):
                new_target = {}
                new_target.update(target)
                new_target["name"] = target["name"] + str(i + 1)
                del new_target["repeat"]
                unfolded_contract["targets"].append(new_target)
        else:
            unfolded_contract["targets"].append(target)

    return unfolded_contract


def get_token(args):
    payload = {'grant_type': 'client_credentials'}
    if "oauth_port" in args and not args.oauth_port is None:
        port = args.oauth_port
    else:
        port = args.port
    url = "http://" + args.host + ":" + str(port) + "/oauth/token"
    print("Getting token from " + url)
    response = requests.post(
        url,
        auth=HTTPBasicAuth(args.oauth_key, args.oauth_secret),
        data=payload)
    print(response.text)
    token = response.json()["access_token"]
    return token


def run_send_feedback(args):
    contract = json.load(open(args.contract, 'r'))
    contract = unfold_contract(contract)
    feature_names = [feature["name"] for feature in contract["features"]]
    response_names = [feature["name"] for feature in contract["targets"]]

    REST_url = "http://" + args.host + ":" + str(args.port) + "/send-feedback"

    for i in range(args.n_requests):
        batch = generate_batch(contract, args.batch_size, 'features')
        response = generate_batch(contract, args.batch_size, 'targets')
        if args.prnt:
            print('-' * 40)
            print("SENDING NEW REQUEST:")

        if not args.grpc:
            REST_request = gen_REST_request(
                batch, features=feature_names, tensor=args.tensor)
            REST_response = gen_REST_request(
                response, features=response_names, tensor=args.tensor)
            reward = 1.0
            REST_feedback = {"request": REST_request,
                             "response": REST_response, "reward": reward}
            if args.prnt:
                print(REST_feedback)

            if args.oauth_key:
                token = get_token(args)
                headers = {'Authorization': 'Bearer ' + token}
                response = requests.post(
                    "http://" + args.host + ":" +
                    str(args.port) + "/api/v0.1/feedback",
                    json=REST_feedback,
                    headers=headers
                )
            else:
                response = requests.post(
                    "http://" + args.host + ":" +
                    str(args.port) + args.ambassador_path +
                    "/api/v0.1/feedback",
                    json=REST_feedback,
                    headers=headers
                )

            if args.prnt:
                print(response)

        elif args.grpc:
            GRPC_request = gen_GRPC_request(
                batch, features=feature_names, tensor=args.tensor)
            GRPC_response = gen_GRPC_request(
                response, features=response_names, tensor=args.tensor)
            reward = 1.0
            GRPC_feedback = prediction_pb2.Feedback(
                request=GRPC_request,
                response=GRPC_response,
                reward=reward
            )

            if args.prnt:
                print(GRPC_feedback)

            channel = grpc.insecure_channel(
                '{}:{}'.format(args.host, args.port))
            stub = prediction_pb2_grpc.SeldonStub(channel)

            if args.oauth_key:
                token = get_token(args)
                metadata = [('oauth_token', token)]
                response = stub.SendFeedback(
                    request=GRPC_feedback, metadata=metadata)
            else:
                response = stub.SendFeedback(request=GRPC_feedback)

            if args.prnt:
                print("RECEIVED RESPONSE:")
                print()


def run_predict(args):
    contract = json.load(open(args.contract, 'r'))
    contract = unfold_contract(contract)
    feature_names = [feature["name"] for feature in contract["features"]]

    REST_url = "http://" + args.host + ":" + str(args.port) + "/predict"

    for i in range(args.n_requests):
        batch = generate_batch(contract, args.batch_size, 'features')
        if args.prnt:
            print('-' * 40)
            print("SENDING NEW REQUEST:")

        if not args.grpc:
            headers = {}
            REST_request = gen_REST_request(
                batch, features=feature_names, tensor=args.tensor)
            if args.prnt:
                print(REST_request)

            if args.oauth_key:
                token = get_token(args)
                headers = {'Authorization': 'Bearer ' + token}
                response = requests.post(
                    "http://" + args.host + ":" +
                    str(args.port) + "/api/v0.1/predictions",
                    json=REST_request,
                    headers=headers
                )
            else:
                response = requests.post(
                    "http://" + args.host + ":" +
                    str(args.port) + args.ambassador_path +
                    "/api/v0.1/predictions",
                    json=REST_request,
                    headers=headers
                )

            jresp = response.json()

            if args.prnt:
                print("RECEIVED RESPONSE:")
                print(jresp)
                print()
        else:
            GRPC_request = gen_GRPC_request(
                batch, features=feature_names, tensor=args.tensor)
            if args.prnt:
                print(GRPC_request)

            channel = grpc.insecure_channel(
                '{}:{}'.format(args.host, args.port))
            stub = prediction_pb2_grpc.SeldonStub(channel)

            if args.oauth_key:
                token = get_token(args)
                metadata = [('oauth_token', token)]
                response = stub.Predict(
                    request=GRPC_request, metadata=metadata)
            else:
                response = stub.Predict(request=GRPC_request)

            if args.prnt:
                print("RECEIVED RESPONSE:")
                print(response)
                print()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("contract", type=str,
                        help="File that contains the data contract")
    parser.add_argument("host", type=str)
    parser.add_argument("port", type=int)
    parser.add_argument("--endpoint", type=str,
                        choices=["predict", "send-feedback"], default="predict")
    parser.add_argument("-b", "--batch-size", type=int, default=1)
    parser.add_argument("-n", "--n-requests", type=int, default=1)
    parser.add_argument("--grpc", action="store_true")
    parser.add_argument("-t", "--tensor", action="store_true")
    parser.add_argument("-p", "--prnt", action="store_true",
                        help="Prints requests and responses")
    parser.add_argument("--oauth-port", type=int)
    parser.add_argument("--oauth-key")
    parser.add_argument("--oauth-secret")
    parser.add_argument("--ambassador-path")

    args = parser.parse_args()

    if args.endpoint == "predict":
        run_predict(args)
    elif args.endpoint == "send-feedback":
        run_send_feedback(args)


if __name__ == "__main__":
    main()
