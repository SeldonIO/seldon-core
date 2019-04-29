import os
import sys, getopt, argparse
import logging
import json

def getDefinition(uri):
    return uri[14:]

def expand(defn,root):
    if "properties" in defn:
        for prop in defn["properties"].keys():
            if "$ref" in defn["properties"][prop]:
                ref = getDefinition(defn["properties"][prop]["$ref"])
                defnNew = expand(root["definitions"][ref],root)
                defn["properties"][prop] = defnNew
            elif "items" in defn["properties"][prop] and "$ref" in defn["properties"][prop]["items"]:
                ref = getDefinition(defn["properties"][prop]["items"]["$ref"])
                defnNew = expand(root["definitions"][ref],root)
                defn["properties"][prop]["items"] = defnNew
    return defn

def simplifyAdditionalProperties(defn):
    if isinstance(defn, dict):
        if "additionalProperties" in defn.keys():
            if isinstance(defn["additionalProperties"], dict):
                if "$ref" in defn["additionalProperties"].keys():
                    defn["additionalProperties"] = True
        for k in defn.keys():
            simplifyAdditionalProperties(defn[k])

"""
Expands the Swagger JSON for the Root Item by expanding out the $ref items to their referring JSON
Needed as CRD OpenAPI validation can't handle $ref elements
"""
if __name__ == '__main__':
    import logging
    logger = logging.getLogger()
    logging.basicConfig(format='%(asctime)s : %(levelname)s : %(name)s : %(message)s', level=logging.DEBUG)
    logger.setLevel(logging.INFO)

    parser = argparse.ArgumentParser(prog='create_replay')
    parser.add_argument('--swagger', help='Swagger OpenAPI file', default="swagger.json")
    parser.add_argument('--root', help='root defn', default="io.k8s.api.core.v1.PodSpec")

    args = parser.parse_args()
    opts = vars(args)

    data = json.load(open(args.swagger))

    root = data["definitions"][args.root]
    expandedRoot = expand(root,data)
    simplifyAdditionalProperties(expandedRoot)
    print(json.dumps(expandedRoot,indent=4))
