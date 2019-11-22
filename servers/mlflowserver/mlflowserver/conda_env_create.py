"""
This script creates a Conda environment from a `conda.yaml`.
The `conda.yaml` file gets fetched from cloud storage.
"""
import os
import logging
import argparse
import json
import yaml
import tempfile
from subprocess import run
from seldon_core.microservice import PARAMETERS_ENV_NAME, parse_parameters
from seldon_core import Storage
from shlex import quote


log = logging.getLogger()
log.setLevel("INFO")

parser = argparse.ArgumentParser()
parser.add_argument(
    "--parameters", type=str, default=os.environ.get(PARAMETERS_ENV_NAME, "[]")
)

# This is already set on the environment_rest and environment_grpc files, but
# we'll define a default just in case.
DEFAULT_CONDA_ENV_NAME = "mlflow"
BASE_REQS_PATH = os.path.join(
    os.path.dirname(os.path.abspath(__file__)), "requirements.txt"
)


def setup_env(model_folder):
    """Sets up a Conda environment.

    This methods creates the Conda environment described by the `MLmodel` file.
    It also injects the minimum set of dependencies described in `base.yaml` on
    this environment.

    Parameters
    --------
    model_folder
        Folder where the MLmodel files are stored.
    """
    mlmodel = read_mlmodel(model_folder)

    flavours = mlmodel["flavors"]
    pyfunc_flavour = flavours["python_function"]
    env_file_name = pyfunc_flavour["env"]
    env_file_path = os.path.join(model_folder, env_file_name)
    env_file_path = inject_base_reqs(env_file_path)

    create_env(env_file_path)


def inject_base_reqs(env_file_path):
    """Injects the base set of requirements.

    Adds a new `pip` entry to in the Conda environment with the base set of
    requirements for `MLFlowServer.py`.
    This list of base packages is defined in `requirements.txt`.

    Parameters
    --------
    env_file_path
        Path to the Conda YAML environment.

    Returns
    -------
    str
        Path to the new YAML environment with the injected base dependencies.
    """
    conda_env = _read_yaml(env_file_path)
    if "dependencies" not in conda_env:
        conda_env["dependencies"] = []

    new_entry = {"pip": [f"-r {BASE_REQS_PATH}"]}
    conda_env["dependencies"].append(new_entry)

    temp_dir = tempfile.mkdtemp()
    new_env_path = os.path.join(temp_dir, "conda.yaml")
    with open(new_env_path, "w") as new_env:
        yaml.dump(conda_env, new_env)

    return new_env_path


def read_mlmodel(model_folder):
    """Reads an MLmodel file.

    Parameters
    ---------
    model_folder
        Folder where the MLmodel files are stored.

    Returns
    --------
    obj
        Dictionary with MLmodel contents.
    """
    log.info(f"Reading MLmodel file")
    mlmodel_path = os.path.join(model_folder, "MLmodel")
    return _read_yaml(mlmodel_path)


def _read_yaml(file_path):
    """Reads a YAML file.

    Parameters
    ---------
    file_path
        Path to the YAML file.

    Returns
    -------
    dict
        Dictionary with YAML file contents.
    """
    with open(file_path, "r") as file:
        return yaml.safe_load(file)


def create_env(env_file_path):
    """Creates Conda environment from YAML.

    Creates a Conda environment from a YAML file describing Python version,
    dependencies, etc.
    The new environment name is read from the `CONDA_ENV_NAME` environment
    variable.
    If the variable is not defined, it falls back to `mlflow`.
    """
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
