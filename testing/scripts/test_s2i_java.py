import os
import time
import pytest

from subprocess import run

from seldon_e2e_utils import initial_rest_request, wait_for_status, wait_for_rollout
from e2e_utils.common import SC_ROOT_PATH
from e2e_utils.models import deploy_model


S2I_IMAGE_BUILD = "seldonio/s2i-java-jni-build"
S2I_IMAGE_RUNTIME = "seldonio/s2i-java-jni-runtime"

S2I_JAVA_VERSION = "0.3.0"

DEFAULT_JAVA_IMAGE = "seldonio/test-s2i-java:0.1.0"
DEFAULT_S2I_FOLDER = os.path.join(SC_ROOT_PATH, "testing", "s2i", "java-jni")


def create_s2i_image(
    s2i_folder=DEFAULT_S2I_FOLDER,
    s2i_version=S2I_JAVA_VERSION,
    image_name=DEFAULT_JAVA_IMAGE,
) -> str:
    cmd = (
        f"s2i build {s2i_folder} "
        f"{S2I_IMAGE_BUILD}:{s2i_version} "
        f"--runtime-image {S2I_IMAGE_RUNTIME}:{s2i_version} "
        f"{image_name}"
    )
    run(cmd, shell=True, check=True)

    return image_name


def kind_load_image(image_name: str):
    cmd = f"kind load docker-image {image_name}"
    run(cmd, shell=True, check=True)


@pytest.mark.sequential
def test_build_s2i_image():
    image_name = create_s2i_image()

    container_name = "jni-model"
    run(
        f"docker run -d --rm --name {container_name} {image_name}",
        shell=True,
        check=True,
    )
    time.sleep(2)
    run(f"docker rm -f {container_name}", shell=True, check=True)


@pytest.mark.sequential
def test_model_rest(namespace):
    image_name = create_s2i_image()
    kind_load_image(image_name)

    deploy_model("mymodel", namespace=namespace, model_image=image_name)
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)

    r = initial_rest_request("mymodel", namespace)

    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["shape"] == [1]
    assert r.json()["data"]["tensor"]["values"] == [1.0]
