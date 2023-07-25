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
"""CLI tool that returns URL of the proxy for particular zone and version."""
import argparse
import functools
import json
import logging
import re
import requests

try:
    unicode
except NameError:
    unicode = str


def urls_for_zone(zone, location_to_urls_map):
    """Returns list of potential proxy URLs for a given zone.

  Returns:
    List of possible URLs, in order of proximity.
  Args:
    zone: GCP zone
    location_to_urls_map: Maps region/country/continent to list of URLs, e.g.:
      {
        "us-west1" : [ us-west1-url ],
        "us-east1" : [ us-east1-url ],
        "us" : [ us-west1-url ],
        ...
      }
  """
    zone_match = re.match("((([a-z]+)-[a-z]+)\d+)-[a-z]", zone)
    if not zone_match:
        raise ValueError("Incorrect zone specified: {}".format(zone))

    # e.g. zone = us-west1-b
    region = zone_match.group(1)  # us-west1
    approx_region = zone_match.group(2)  # us-west
    country = zone_match.group(3)  # us

    urls = []
    if region in location_to_urls_map:
        urls.extend([
            url for url in location_to_urls_map[region] if url not in urls
        ])

    region_regex = re.compile("([a-z]+-[a-z]+)\d+")
    for location in location_to_urls_map:
        region_match = region_regex.match(location)
        if region_match and region_match.group(1) == approx_region:
            urls.extend([
                url for url in location_to_urls_map[location] if url not in urls
            ])

    if country in location_to_urls_map:
        urls.extend([
            url for url in location_to_urls_map[country] if url not in urls
        ])

    return urls


def main():
    unicode_type = functools.partial(unicode, encoding="utf8")
    parser = argparse.ArgumentParser(description="Get proxy URL")
    parser.add_argument("--config-file-path", required=True, type=unicode_type)
    parser.add_argument("--location", required=True, type=unicode_type)
    parser.add_argument("--version", required=True, type=unicode_type)

    args = parser.parse_args()
    with open(args.config_file_path, "r") as config_file:
        data = json.loads(config_file.read())

    agent_containers_config = data["agent-docker-containers"]
    version = args.version
    if version not in agent_containers_config:
        version = "latest"
    if version not in agent_containers_config:
        raise ValueError("Version latest not found in the config file.")
    container_config = agent_containers_config[version]
    regional_urls = container_config["proxy-urls"]

    location = args.location
    urls = urls_for_zone(location, regional_urls)
    if not urls:
        raise ValueError("No valid URLs found for zone: {}".format(location))

    for url in urls:
        try:
            status_code = requests.head(url).status_code
        except requests.ConnectionError:
            pass
        expected_codes = frozenset([307])
        # 307 - Temporary Redirect, Proxy server sends this if VM has access rights.
        if status_code in expected_codes:
            logging.debug("Status code from the url %s", status_code)
            print(url)
            exit(0)
        logging.debug(
            "Incorrect status_code from the server: %s. Expected: %s",
            status_code, expected_codes
        )
    raise ValueError("No working URL found")


if __name__ == '__main__':
    main()
