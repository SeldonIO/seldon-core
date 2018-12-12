import argparse
import numpy as np
import json
import requests
import urllib
from google.protobuf.struct_pb2 import ListValue
import grpc
from time import time

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
    return np.array(values)[vals]


def generate_batch(contract, n, field):
    feature_batches = []
    ty_set = set()
    for feature_def in contract[field]:
        ty_set.add(feature_def["ftype"])
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
            batch = np.around(batch, decimals=3)
            batch = reconciliate_cont_type(batch, feature_def["dtype"])
        elif feature_def["ftype"] == "categorical":
            batch = gen_categorical(feature_def["values"], [n, 1])
        feature_batches.append(batch)
    if len(ty_set) == 1:
        return np.concatenate(feature_batches, axis=1)
    else:
        out = np.empty((n, len(contract['features'])), dtype=object)
        return np.concatenate(feature_batches, axis=1, out=out)


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
                new_target["name"] = target["name"] + ":" + str(i)
                del new_target["repeat"]
                unfolded_contract["targets"].append(new_target)
        else:
            unfolded_contract["targets"].append(target)

    return unfolded_contract


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

        if not args.grpc and not args.fbs:
            REST_request = gen_REST_request(
                batch, features=feature_names, tensor=args.tensor)
            REST_response = gen_REST_request(
                response, features=response_names, tensor=args.tensor)
            reward = 1.0
            REST_feedback = {"request": REST_request,
                             "response": REST_response, "reward": reward}
            if args.prnt:
                print(REST_feedback)

            t1 = time()
            response = requests.post(
                REST_url,
                data={"json": json.dumps(REST_feedback)})
            t2 = time()

            if args.prnt:
                print("Time " + str(t2 - t1))
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
            stub = prediction_pb2_grpc.ModelStub(channel)
            response = stub.SendFeedback(GRPC_feedback)

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

        if not args.grpc and not args.fbs:
            REST_request = gen_REST_request(
                batch, features=feature_names, tensor=args.tensor)
            if args.prnt:
                print(REST_request)

            t1 = time()
            response = requests.post(
                REST_url,
                data={"json": json.dumps(REST_request), "isDefault": True})
            t2 = time()
            jresp = response.json()

            if args.prnt:
                print("RECEIVED RESPONSE:")
                print(jresp)
                print()
                print("Time " + str(t2 - t1))
        elif args.grpc:
            GRPC_request = gen_GRPC_request(
                batch, features=feature_names, tensor=args.tensor)
            if args.prnt:
                print(GRPC_request)

            channel = grpc.insecure_channel(
                '{}:{}'.format(args.host, args.port))
            stub = prediction_pb2_grpc.ModelStub(channel)
            response = stub.Predict(GRPC_request)

            if args.prnt:
                print("RECEIVED RESPONSE:")
                print(response)
                print()
        elif args.fbs:
            import socket
            import struct
            from seldon_core.tester_flatbuffers import NumpyArrayToSeldonRPC, SeldonRPCToNumpyArray
            data = NumpyArrayToSeldonRPC(batch, feature_names)
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.connect((args.host, args.port))
            totalsent = 0
            MSGLEN = len(data)
            print("Will send", MSGLEN, "bytes")
            while totalsent < MSGLEN:
                sent = s.send(data[totalsent:])
                if sent == 0:
                    raise RuntimeError("socket connection broken")
                totalsent = totalsent + sent
            data = s.recv(4)
            obj = struct.unpack('<i', data)
            len_msg = obj[0]
            data = s.recv(len_msg)
            arr = SeldonRPCToNumpyArray(data)
            print(arr)


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
    parser.add_argument("--fbs", action="store_true")
    parser.add_argument("-t", "--tensor", action="store_true")
    parser.add_argument("-p", "--prnt", action="store_true",
                        help="Prints requests and responses")

    args = parser.parse_args()

    if args.endpoint == "predict":
        run_predict(args)
    elif args.endpoint == "send-feedback":
        run_send_feedback(args)


if __name__ == "__main__":
    main()
