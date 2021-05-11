# Copyright 2020 kubeflow.org.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Copied from kfserving project as starter.
#


import glob
import logging
import os
import tempfile

_LOCAL_PREFIX = "file://"

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
                "rclone binary not found - rclone-based storage funcionalty disabled"
            )

        if dest is None:
            dest = tempfile.mkdtemp()

        kwargs = {}
        if self.cfg_file is not None:
            kwargs["config"] = os.path.abspath(os.path.expanduser(self.cfg_file))

        rclone.copy(src, dest, **kwargs)
        return dest


class Storage:
    @staticmethod
    def download(uri: str, out_dir: str = None) -> str:
        logging.info("Copying contents of %s to local", uri)

        is_local = False
        if uri.startswith(_LOCAL_PREFIX) or os.path.exists(uri):
            is_local = True

        if out_dir is None:
            if is_local:
                # noop if out_dir is not set and the path is local
                return Storage._download_local(uri)
            out_dir = tempfile.mkdtemp()

        if is_local:
            return Storage._download_local(uri, out_dir)
        else:
            raise Exception(
                f"Cannot recognize storage type for {uri}. "
                f"Only {_LOCAL_PREFIX} is currently available as storage type."
            )

        logging.info("Successfully copied %s to %s", uri, out_dir)
        return out_dir

    @staticmethod
    def _download_local(uri: str, out_dir: str = None) -> str:
        local_path = uri.replace(_LOCAL_PREFIX, "", 1)
        if not os.path.exists(local_path):
            raise RuntimeError("Local path %s does not exist." % (uri))

        if out_dir is None:
            return local_path
        elif not os.path.isdir(out_dir):
            os.makedirs(out_dir)

        if os.path.isdir(local_path):
            local_path = os.path.join(local_path, "*")

        for src in glob.glob(local_path):
            _, tail = os.path.split(src)
            dest_path = os.path.join(out_dir, tail)
            logging.info("Linking: %s to %s", src, dest_path)
            os.symlink(src, dest_path)
        return out_dir
