# Copyright 2019 kubeflow.org.
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

import pytest
import seldon_core


STORAGE_MODULE = "seldon_core.storage"


def test_storage_local_path():
    abs_path = "file:///"
    relative_path = "file://."
    assert seldon_core.Storage.download(abs_path) == abs_path.replace("file://", "", 1)
    assert seldon_core.Storage.download(relative_path) == relative_path.replace(
        "file://", "", 1
    )


@pytest.mark.parametrize(
    "uri",
    [
        "file:///not/existing/path",
        "s3://foo/bar",
        "gs://foo/bar",
        "https://accountname.blob.core.windows.net/container/some/blob/",
    ],
)
def test_invalid_storage_uri(uri):
    with pytest.raises(Exception):
        seldon_core.Storage.download(uri)


def test_no_prefix_local_path():
    abs_path = "/"
    relative_path = "."
    assert seldon_core.Storage.download(abs_path) == abs_path
    assert seldon_core.Storage.download(relative_path) == relative_path
