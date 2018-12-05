#!/usr/bin/env python

import requests
from requests.auth import HTTPBasicAuth
import pprint

try:
    from commands import getoutput # python 2
except ImportError:
    from subprocess import getoutput # python 3

NAMESPACE="seldon"
SELDON_API_IP="localhost"

def pp(o):
    pprinter = pprint.PrettyPrinter(indent=4)
    pprinter.pprint(o)

def get_token():
    payload = {'grant_type': 'client_credentials'}
    url="http://{}:8080/oauth/token".format(SELDON_API_IP)
    response = requests.post(
                url,
                auth=HTTPBasicAuth('oauth-key', 'oauth-secret'),
                data=payload)
    token =  response.json()["access_token"]
    return token

def rest_request():
    token = get_token()
    headers = {'Authorization': 'Bearer '+token}
    payload = {"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[0,0,1,1]}}}
    response = requests.post(
                "http://{}:8080/api/v0.1/predictions".format(SELDON_API_IP),
                headers=headers,
                json=payload)
    print(response.text)
    
def main():
    rest_request()

if __name__ == "__main__":
    main()

