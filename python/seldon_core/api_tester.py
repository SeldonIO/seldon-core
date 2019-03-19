import argparse
import json
from seldon_core.seldon_client import SeldonClient
from seldon_core.microservice_tester import unfold_contract, generate_batch
import logging


def get_seldon_client(args) -> SeldonClient:
    """
    Get the appropriate Seldon Client based on args

    Parameters
    ----------
    args
       Command line args


    Returns
    -------
       A Seldon Client

    """
    endpoint = args.host + ":" + str(args.port)
    if args.oauth_key:
        if args.grpc:
            seldon_grpc_endpoint = endpoint
            seldon_rest_endpoint = args.host + ":" + str(args.oauth_port)
        else:
            seldon_grpc_endpoint = None
            seldon_rest_endpoint = endpoint
        sc = SeldonClient(gateway="seldon", seldon_rest_endpoint=seldon_rest_endpoint,
                          seldon_grpc_endpoint=seldon_grpc_endpoint,
                          oauth_key=args.oauth_key, oauth_secret=args.oauth_secret)
    else:
        ambassador_endpoint = endpoint
        sc = SeldonClient(gateway="ambassador", ambassador_endpoint=ambassador_endpoint,
                          deployment_name=args.deployment, namespace=args.namespace)
    return sc


def run_send_feedback(args):
    """
    Do a semd-feedback call to the Seldon API

    Parameters
    ----------
    args
       Command line args

    """
    contract = json.load(open(args.contract, 'r'))
    contract = unfold_contract(contract)
    sc = get_seldon_client(args)
    if args.grpc:
        transport = "grpc"
    else:
        transport = "rest"

    for i in range(args.n_requests):
        batch = generate_batch(contract, args.batch_size, 'features')
        response_predict = sc.predict(data=batch, deployment_name=args.deployment)
        response_feedback = sc.feedback(prediction_request=response_predict.request,
                                        prediction_response=response_predict.response, reward=1.0,
                                        deployment_name=args.deployment, transport=transport)
        if args.prnt:
            print(f"RECEIVED RESPONSE:\n{response_feedback}\n")


def run_predict(args):
    """
    Make a prediction call to the Seldon API

    Parameters
    ----------
    args
       Command line args

    """
    contract = json.load(open(args.contract, 'r'))
    contract = unfold_contract(contract)
    feature_names = [feature["name"] for feature in contract["features"]]

    sc = get_seldon_client(args)
    if args.grpc:
        transport = "grpc"
    else:
        transport = "rest"
    payload_type = "tensor" if args.tensor else "ndarray"

    for i in range(args.n_requests):
        batch = generate_batch(contract, args.batch_size, 'features')
        if args.prnt:
            print(f"{'-' * 40}\nSENDING NEW REQUEST:\n")
            print(batch)
        response_predict = sc.predict(data=batch, deployment_name=args.deployment, names=feature_names, payload_type=payload_type)
        if args.prnt:
            print(f"RECEIVED RESPONSE:\n{response_predict.response}\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("contract", type=str,
                        help="File that contains the data contract")
    parser.add_argument("host", type=str)
    parser.add_argument("port", type=int)
    parser.add_argument("deployment", type=str, nargs='?', default="mymodel")
    parser.add_argument("--endpoint", type=str, choices=["predict", "send-feedback"], default="predict")
    parser.add_argument("-b", "--batch-size", type=int, default=1)
    parser.add_argument("-n", "--n-requests", type=int, default=1)
    parser.add_argument("--grpc", action="store_true")
    parser.add_argument("-t", "--tensor", action="store_true")
    parser.add_argument("-p", "--prnt", action="store_true", help="Prints requests and responses")
    parser.add_argument("--log-level", type=str, choices=["DEBUG", "INFO", "ERROR"], default="ERROR")
    parser.add_argument("--namespace", type=str)
    parser.add_argument("--oauth-port", type=int)
    parser.add_argument("--oauth-key")
    parser.add_argument("--oauth-secret")

    args = parser.parse_args()
    LOG_FORMAT = '%(asctime)s - %(name)s:%(funcName)s:%(lineno)s - %(levelname)s:  %(message)s'
    if args.log_level == "DEBUG":
        log_level = logging.DEBUG
    elif args.log_level == "INFO":
        log_level = logging.INFO
    else:
        log_level = logging.ERROR
    logging.basicConfig(level=log_level, format=LOG_FORMAT)

    if args.endpoint == "predict":
        run_predict(args)
    elif args.endpoint == "send-feedback":
        run_send_feedback(args)


if __name__ == "__main__":
    main()
