#!/usr/bin/env python

import requests
import json

def rest_request_ambassador(deploymentName, request, endpoint="localhost:8003"):
    response = requests.post(
        "http://" + endpoint + "/seldon/" + deploymentName + "/api/v0.1/predictions",
        json=request)
    return response.json()

def rest_request():
    payload = {"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}
    response_dict=rest_request_ambassador("seldon-deployment-example", payload, endpoint="localhost:8080")
    response_json=json.dumps(response_dict, sort_keys=True, indent=4, separators=(',', ': '))
    print(response_json)

def main():
    rest_request()

if __name__ == "__main__":
    main()

