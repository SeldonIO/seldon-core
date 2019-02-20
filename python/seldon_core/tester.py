import argparse
import numpy as np
import json
from typing import Dict, List, Union
from numpy.core.multiarray import ndarray
from seldon_core.seldon_client import SeldonClient


def gen_continuous(range: List, n: int) -> np.ndarray:
    if range[0] == "inf" and range[1] == "inf":
        return np.random.normal(size=n)
    if range[0] == "inf":
        return range[1] - np.random.lognormal(size=n)
    if range[1] == "inf":
        return range[0] + np.random.lognormal(size=n)
    return np.random.uniform(range[0], range[1], size=n)


def reconciliate_cont_type(feature: np.ndarray, dtype) -> np.ndarray:
    if dtype == "FLOAT":
        return feature
    if dtype == "INT":
        return (feature + 0.5).astype(int).astype(float)


def gen_categorical(values:List[str], n: List[int]) -> np.ndarray:
    vals = np.random.randint(len(values), size=n)
    return np.array(values)[vals]


def generate_batch(contract: Dict, n: int, field: str) -> np.ndarray:
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


def unfold_contract(contract: Dict) -> Dict:
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
    endpoint = args.host + ":" + str(args.port)
    sc = SeldonClient(microservice_endpoint=endpoint)

    for i in range(args.n_requests):
        batch = generate_batch(contract, args.batch_size, 'features')
        if args.prnt:
            print('-' * 40)
            print("SENDING NEW REQUEST:")

        if not args.grpc:
            response_predict = sc.microservice(data=batch, transport="rest")
            response_feedback = sc.microservice_feedback(prediction_request=response_predict.request,prediction_response=response_predict.response,reward=1.0)
            if args.prnt:
                print("RECEIVED RESPONSE:")
                print(response_feedback)
                print()

        elif args.grpc:
            response_predict = sc.microservice(data=batch, transport="grpc")
            response_feedback = sc.microservice_feedback(prediction_request=response_predict.request,prediction_response=response_predict.response,reward=1.0)
            if args.prnt:
                print("RECEIVED RESPONSE:")
                print(response_feedback)
                print()


def run_predict(args):
    contract = json.load(open(args.contract, 'r'))
    contract = unfold_contract(contract)
    feature_names = [feature["name"] for feature in contract["features"]]

    endpoint = args.host + ":" + str(args.port)
    sc = SeldonClient(microservice_endpoint=endpoint)

    for i in range(args.n_requests):
        batch: ndarray = generate_batch(contract, args.batch_size, 'features')
        if args.prnt:
            print('-' * 40)
            print("SENDING NEW REQUEST:")

        if not args.grpc:
            response = sc.microservice(data=batch,transport="rest",method="predict")

            if args.prnt:
                print("RECEIVED RESPONSE:")
                print(response)
                print()
        elif args.grpc:
            response = sc.microservice(data=batch,transport="grpc",method="predict")
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

    args = parser.parse_args()

    if args.endpoint == "predict":
        run_predict(args)
    elif args.endpoint == "send-feedback":
        run_send_feedback(args)


if __name__ == "__main__":
    main()
