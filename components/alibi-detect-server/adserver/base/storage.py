import os
import sys
import logging
import tempfile
from distutils.util import strtobool

ARTIFACT_DOWNLOAD_LOCATION = os.environ.get("DRIFT_ARTIFACTS_DIR", "/tmp")

logger = logging.getLogger(__name__)

try:
    from sh import rclone
except ImportError:
    logger.warning(
        "rclone-based storage funcionality not available without rclone binary"
    )
    rclone = None


class Rclone:
    def __init__(self, cfg_file: str = None):
        self.cfg_file = cfg_file

    def copy(self, src: str, dest: str = None):
        if rclone is None:
            raise RuntimeError(
                "rclone binary not found - rclone-based storage funcionality disabled"
            )

        if dest is None:
            dest = tempfile.mkdtemp(dir=ARTIFACT_DOWNLOAD_LOCATION)

        args = ["-vv"]
        kwargs = {}
        if self.cfg_file is not None:
            kwargs["config"] = os.path.abspath(os.path.expanduser(self.cfg_file))

        rclone.copy(src, dest, *args, **kwargs, _out=sys.stdout, _err_to_out=True)
        return dest


def download_model(storage_uri) -> str:
    return Rclone().copy(storage_uri)
