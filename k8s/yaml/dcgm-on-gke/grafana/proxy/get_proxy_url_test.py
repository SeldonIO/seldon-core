#!/usr/bin/env python3
# Copyright 2019 The Kubeflow Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import unittest

from get_proxy_url import urls_for_zone

url_map_json = """
    {
      "us": ["https://datalab-us-west1.cloud.google.com"],
      "us-west1": ["https://datalab-us-west1.cloud.google.com"],
      "us-west2": ["https://datalab-us-west2.cloud.google.com"],
      "us-east1": ["https://datalab-us-east1.cloud.google.com"]
    }
    """


class TestUrlsForZone(unittest.TestCase):

    def test_get_urls(self):
        self.assertEqual([
            "https://datalab-us-east1.cloud.google.com",
            "https://datalab-us-west1.cloud.google.com"
        ], urls_for_zone("us-east1-a", json.loads(url_map_json)))

    def test_get_urls_no_match(self):
        self.assertEqual([],
                         urls_for_zone(
                             "euro-west1-a", json.loads(url_map_json)
                         ))

    def test_get_urls_incorrect_format(self):
        with self.assertRaises(ValueError):
            urls_for_zone("weird-format-a", json.loads(url_map_json))

    def test_get_urls_priority(self):
        self.assertEqual([
            "https://datalab-us-west1.cloud.google.com",
            "https://datalab-us-west2.cloud.google.com"
        ], urls_for_zone("us-west1-a", json.loads(url_map_json)))


if __name__ == '__main__':
    unittest.main()
