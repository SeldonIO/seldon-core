import os
import kfserving
import logging
import tempfile
from distutils.util import strtobool


logger = logging.getLogger(__name__)

try:
    from sh import rclone
except ImportError:
    logger.warning(
        "rclone-based storage funcionality not available without rclone binary"
    )
    rclone = None


def _rclone_enabled():
    # IF RCLONE_ENABLED variable set explicitly we read from it
    enabled = os.environ.get("RCLONE_ENABLED", None)
    if enabled is not None:
        return strtobool(enabled)

    # Otherwise we determine if Rclone config is provided
    for key in os.environ.keys():
        if "RCLONE_CONFIG" in key:
            return True
    else:
        return False


RCLONE_ENABLED = _rclone_enabled()


class Rclone:
    def __init__(self, cfg_file: str = None):
        self.cfg_file = cfg_file

    def copy(self, src: str, dest: str = None):
        if rclone is None:
            raise RuntimeError(
                "rclone binary not found - rclone-based storage funcionality disabled"
            )

        if dest is None:
            dest = tempfile.mkdtemp()

        kwargs = {}
        if self.cfg_file is not None:
            kwargs["config"] = os.path.abspath(os.path.expanduser(self.cfg_file))

        rclone.copy(src, dest, **kwargs)
        return dest


def download_model(storage_uri) -> str:
    if RCLONE_ENABLED:
        return Rclone().copy(storage_uri)
    else:
        return kfserving.Storage.download(storage_uri)
