import os
import sys, getopt, argparse
import logging
from random import randint,random
import json
from kazoo.client import KazooClient
import seldonengine_pb2 as spb
from google.protobuf import json_format
import base64

def is_json_data(data):
    if (data != None) and (len(data)>0):
        return data[0] == '{' or data[0] == '['
    else:
        return False


def json_compress(json_data):
    d = json.loads(json_data)
    return json.dumps(d, sort_keys=True, separators=(',',':'))

def node_set(zk_client, node_path, node_value):
    if is_json_data(node_value):
        node_value = json_compress(node_value)
    node_value = node_value.strip() if node_value != None else node_value

    if zk_client.exists(node_path):
        retVal = zk_client.set(node_path,node_value)
    else:
        retVal = zk_client.create(node_path,node_value,makepath=True)
    print "updated zk node[{node_path}]".format(node_path=node_path)



if __name__ == '__main__':
    import logging
    logger = logging.getLogger()
    logging.basicConfig(format='%(asctime)s : %(levelname)s : %(name)s : %(message)s', level=logging.DEBUG)
    logger.setLevel(logging.INFO)

    parser = argparse.ArgumentParser(prog='create_replay')
    parser.add_argument('--zookeeper', help='zookeeper host:port', default="0.0.0.0:2181")
    parser.add_argument('--json', help='deployment json', required=True)
    #parser.add_argument('--deployment-store-folder', help='deployment store folder', default="")

    args = parser.parse_args()
    opts = vars(args)

    zk_client = KazooClient(hosts=args.zookeeper)
    zk_client.start()

    with open(args.json) as json_data:
        d = json.load(json_data)

    # check json parses
    deployment = spb.DeploymentDef()
    json_format.ParseDict(d,deployment)
    json_string = json_format.MessageToJson(deployment)

    node_set(zk_client,"/deployments/"+d["id"],json_string)


    json_string = json_format.MessageToJson(deployment.predictor)
    x = base64.b64encode(json_string)
    print "Base64 encoded predictor for Cluster Manager"
    print x



