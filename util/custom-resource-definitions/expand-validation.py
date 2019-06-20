import os
import sys, getopt, argparse
import logging
import json

def getDefinition(uri):
    return uri[14:]

def expand(defn,root):
    if "properties" in defn:
        badProps = []
        for prop in defn["properties"].keys():
            if "$ref" in defn["properties"][prop]:
                ref = getDefinition(defn["properties"][prop]["$ref"])
                defnNew = expand(root["definitions"][ref],root)
                if defnNew:
                    defn["properties"][prop] = defnNew
                else:
                    badProps.append(prop)
            elif "items" in defn["properties"][prop] and "$ref" in defn["properties"][prop]["items"]:
                ref = getDefinition(defn["properties"][prop]["items"]["$ref"])
                defnNew = expand(root["definitions"][ref],root)
                if defnNew:
                    defn["properties"][prop]["items"] = defnNew
                else:
                    badProps.append(prop)
            ## Temporary hack until https://github.com/go-openapi/validate/issues/108 is fixed
            ## Causes failure of CRD validation with properties with "items" property
            if "items" in defn["properties"] and "items" in defn["properties"]["items"]:
                return None
        for prop in badProps:
            del defn["properties"][prop]
    return defn

def simplifyAdditionalProperties(defn):
    if isinstance(defn, dict):
        if "additionalProperties" in defn.keys():
            if isinstance(defn["additionalProperties"], dict):
                if "$ref" in defn["additionalProperties"].keys():
                    del defn["additionalProperties"]
                    #defn["additionalProperties"] = True
        for k in defn.keys():
            simplifyAdditionalProperties(defn[k])

def removeDescriptions(defn):
    if isinstance(defn, dict):
        if "description" in defn.keys():
            del defn["description"]
        for k in defn.keys():
            removeDescriptions(defn[k])

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
    removeDescriptions(expandedRoot)
    print(json.dumps(expandedRoot,indent=4))
