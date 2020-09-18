import os
import time

from subprocess import run

from e2e_utils.common import SC_ROOT_PATH


S2I_IMAGE_BUILD = "seldonio/s2i-java-jni-build"
S2I_IMAGE_RUNTIME = "seldonio/s2i-java-jni-runtime"

S2I_JAVA_VERSION = "0.3.0"

DEFAULT_JAVA_IMAGE = "seldonio/test-s2i-java:0.1.0"
DEFAULT_S2I_FOLDER = os.path.join(SC_ROOT_PATH, "testing", "s2i", "java-jni")


def create_s2i_image(
    s2i_folder=DEFAULT_S2I_FOLDER,
    s2i_version=S2I_JAVA_VERSION,
    image_name=DEFAULT_JAVA_IMAGE,
):
    cmd = (
        f"s2i build {s2i_folder} "
        f"{S2I_IMAGE_BUILD}:{s2i_version} "
        f"--runtime-image {S2I_IMAGE_RUNTIME}:{s2i_version} "
        f"{image_name}"
    )
    run(cmd, shell=True, check=True)


def test_build_s2i_image():
    create_s2i_image()

    container_name = "jni-model"
    run(
        f"docker run -d --rm --name {container_name} {DEFAULT_JAVA_IMAGE}",
        shell=True,
        check=True,
    )
    time.sleep(2)
    run(f"docker rm -f {container_name}", shell=True, check=True)
