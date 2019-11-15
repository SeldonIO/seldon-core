"""
This script creates a Conda environment from a `conda.yaml`.
The `conda.yaml` file gets fetched from cloud storage.
"""
import os
import logging
import argparse
import json
import yaml
from subprocess import run
from seldon_core.microservice import PARAMETERS_ENV_NAME, parse_parameters
from seldon_core import Storage
from shlex import quote


log = logging.getLogger()

parser = argparse.ArgumentParser()
parser.add_argument(
    "--parameters", type=str, default=os.environ.get(PARAMETERS_ENV_NAME, "[]")
)

# This is already set on the environment_rest and environment_grpc files, but
# we'll define a default just in case.
DEFAULT_CONDA_ENV_NAME = "mlflow"


def setup_env(model_folder):
    mlmodel = read_mlmodel(model_folder)

    flavours = mlmodel["flavors"]
    pyfunc_flavour = flavours["python_function"]
    env_file_name = pyfunc_flavour["env"]
    env_file_path = os.path.join(model_folder, env_file_name)

    create_env(env_file_path)


def read_mlmodel(model_folder):
    log.info(f"Reading MLmodel file")
    mlmodel_path = os.path.join(model_folder, "MLmodel")
    with open(mlmodel_path, "r") as mlmodel_file:
        return yaml.safe_load(mlmodel_file)


def create_env(env_file_path):
    env_file_name = os.path.basename(env_file_path)
    env_name = os.getenv("CONDA_ENV_NAME", DEFAULT_CONDA_ENV_NAME)
    env_name = quote(env_name)
    env_file_path = quote(env_file_path)

    log.info(f"Creating Conda environment '{env_name}' from {env_file_name}")

    cmd = f"conda env create -n {env_name} --file {env_file_path}"
    run(cmd, shell=True, check=True)


def main(args):
    parameters = parse_parameters(json.loads(args.parameters))
    model_uri = parameters["model_uri"]

    # TODO: Cache downloaded model for MLFlowServer.py
    log.info(f"Downloading model from {model_uri}")
    model_folder = Storage.download(model_uri)
    setup_env(model_folder)


if __name__ == "__main__":
    args = parser.parse_args()
    main(args)
