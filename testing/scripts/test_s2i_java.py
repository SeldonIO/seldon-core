import os
import time
import pytest

from subprocess import run

from seldon_e2e_utils import initial_rest_request, wait_for_status, wait_for_rollout
from e2e_utils.common import SC_ROOT_PATH
from e2e_utils.models import deploy_model

JAVA_S2I_FOLDER = os.path.join(SC_ROOT_PATH, "testing", "s2i", "java-jni")

S2I_JNI_PARAMETERS = {
    "s2i_folder": JAVA_S2I_FOLDER,
    "s2i_image": "seldonio/s2i-java-jni-build:0.3.0",
    "image_name": "seldonio/test-s2i-java-jni:0.1.0",
    "s2i_runtime_image": "seldonio/s2i-java-jni-runtime:0.3.0",
}

S2I_JAVA_PARAMETERS = {
    "s2i_folder": JAVA_S2I_FOLDER,
    "s2i_image": "seldonio/seldon-core-s2i-java-build:0.2",
    "image_name": "seldonio/test-s2i-java:0.1.0",
    "s2i_runtime_image": "seldonio/seldon-core-s2i-java-runtime:0.2",
}


@pytest.mark.sequential
@pytest.mark.parametrize(
    "s2i_image", [S2I_JAVA_PARAMETERS, S2I_JNI_PARAMETERS], indirect=True,
)
def test_build_s2i_image(s2i_image):
    container_name = "jni-model"
    run(
        f"docker run -d --rm --name {container_name} {s2i_image}",
        shell=True,
        check=True,
    )
    time.sleep(2)
    run(f"docker rm -f {container_name}", shell=True, check=True)


@pytest.mark.sequential
@pytest.mark.parametrize(
    "s2i_image", [S2I_JAVA_PARAMETERS, S2I_JNI_PARAMETERS], indirect=True,
)
def test_model_rest(s2i_image, namespace):
    deploy_model("mymodel", namespace=namespace, model_image=s2i_image)
    wait_for_status("mymodel", namespace)
    wait_for_rollout("mymodel", namespace)

    r = initial_rest_request("mymodel", namespace)

    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["shape"] == [1]
    assert r.json()["data"]["tensor"]["values"] == [1.0]
