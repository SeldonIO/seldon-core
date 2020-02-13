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
from minio import Minio, error
import unittest.mock as mock
from .utils import skipif_gcs_missing
from seldon_core.imports_helper import _GCS_PRESENT

if _GCS_PRESENT:
    from google.cloud import exceptions

STORAGE_MODULE = "seldon_core.storage"


def test_storage_local_path():
    abs_path = "file:///"
    relative_path = "file://."
    assert seldon_core.Storage.download(abs_path) == abs_path.replace("file://", "", 1)
    assert seldon_core.Storage.download(relative_path) == relative_path.replace(
        "file://", "", 1
    )


def test_storage_local_path_exception():
    not_exist_path = "file:///some/random/path"
    with pytest.raises(Exception):
        seldon_core.Storage.download(not_exist_path)


def test_no_prefix_local_path():
    abs_path = "/"
    relative_path = "."
    assert seldon_core.Storage.download(abs_path) == abs_path
    assert seldon_core.Storage.download(relative_path) == relative_path


@skipif_gcs_missing
@mock.patch(STORAGE_MODULE + ".storage")
def test_mock_gcs(mock_storage):
    gcs_path = "gs://foo/bar"
    mock_obj = mock.MagicMock()
    mock_obj.name = "mock.object"
    mock_storage.Client().bucket().list_blobs().__iter__.return_value = [mock_obj]
    assert seldon_core.Storage.download(gcs_path)


def test_storage_blob_exception():
    blob_path = "https://accountname.blob.core.windows.net/container/some/blob/"
    with pytest.raises(Exception):
        seldon_core.Storage.download(blob_path)


@mock.patch("urllib3.PoolManager")
@mock.patch(STORAGE_MODULE + ".Minio")
def test_storage_s3_exception(mock_connection, mock_minio):
    minio_path = "s3://foo/bar"
    # Create mock connection
    mock_server = mock.MagicMock()
    mock_connection.return_value = mock_server
    # Create mock client
    mock_minio.return_value = Minio(
        "s3.us.cloud-object-storage.appdomain.cloud", secure=True
    )
    with pytest.raises(Exception):
        seldon_core.Storage.download(minio_path)


@mock.patch("urllib3.PoolManager")
@mock.patch(STORAGE_MODULE + ".Minio")
def test_no_permission_buckets(mock_connection, mock_minio):
    bad_s3_path = "s3://random/path"
    # Access private buckets without credentials
    mock_minio.return_value = Minio(
        "s3.us.cloud-object-storage.appdomain.cloud", secure=True
    )
    mock_connection.side_effect = error.AccessDenied(None)
    with pytest.raises(error.AccessDenied):
        seldon_core.Storage.download(bad_s3_path)


@skipif_gcs_missing
@mock.patch("urllib3.PoolManager")
def test_no_permission_buckets_gcs(mock_connection):
    bad_gcs_path = "gs://random/path"

    mock_connection.side_effect = exceptions.Forbidden(None)
    with pytest.raises(exceptions.Forbidden):
        seldon_core.Storage.download(bad_gcs_path)
