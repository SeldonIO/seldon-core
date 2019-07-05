import json
import yaml
import logging
import argparse

def removeDescriptions(defn):
    if isinstance(defn, dict):
        if "description" in defn.keys():
            del defn["description"]
        for k in defn.keys():
            removeDescriptions(defn[k])

if __name__ == '__main__':
    import logging
    logger = logging.getLogger()
    logging.basicConfig(format='%(asctime)s : %(levelname)s : %(name)s : %(message)s', level=logging.DEBUG)
    logger.setLevel(logging.INFO)

    parser = argparse.ArgumentParser(prog='create_replay')
    parser.add_argument('--crd', help='CRD Json file', default="crd.tpl.json")
    parser.add_argument('--object-meta', help='ObjectMeta Json file', default="object-meta.json")
    parser.add_argument('--pod-spec', help='PodSpec Json file', default="pod-spec.json")
    parser.add_argument('--hpa-spec', help='HpaSpec Json file', default="hpa-spec.json")

    args = parser.parse_args()
    opts = vars(args)

    crd = json.load(open(args.crd))
    object_meta = json.load(open(args.object_meta))
    pod_spec = json.load(open(args.pod_spec))
    hpa_spec = json.load(open(args.hpa_spec))
    crd["spec"]["validation"]["openAPIV3Schema"]["properties"]["spec"]["properties"]["predictors"]["items"][
        "properties"]["componentSpecs"]["items"]["properties"]["metadata"] = object_meta
    crd["spec"]["validation"]["openAPIV3Schema"]["properties"]["spec"]["properties"]["predictors"]["items"][
        "properties"]["componentSpecs"]["items"]["properties"]["spec"] = pod_spec
    crd["spec"]["validation"]["openAPIV3Schema"]["properties"]["spec"]["properties"]["predictors"]["items"][
        "properties"]["componentSpecs"]["items"]["properties"]["hpaSpec"]["properties"]["metrics"]["items"] = hpa_spec
    removeDescriptions(crd)
    print(yaml.dump(crd))