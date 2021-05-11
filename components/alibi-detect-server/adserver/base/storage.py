import os
import kfserving
from distutils.util import strtobool
from seldon_core.storage import Rclone


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


def download_model(storage_uri) -> str:
    if RCLONE_ENABLED:
        return Rclone().copy(storage_uri)
    else:
        return kfserving.Storage.download(storage_uri)
