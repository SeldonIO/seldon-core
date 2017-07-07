import os
import sys, getopt, argparse
import logging
from random import randint,random
import json
from kazoo.client import KazooClient
import seldonengine_pb2 as spb
from google.protobuf import json_format

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
    parser.add_argument('--service-host', help='service endpoint', default="0.0.0.0")
    parser.add_argument('--service-port', help='service port', type=int, default=5000)
    parser.add_argument('--oauth-key', help='oauth key', default="client")
    parser.add_argument('--oauth-secret', help='oauth secret', default="secret")
    parser.add_argument('--deployment-id', help='deployment id', default="sd-1")

    args = parser.parse_args()
    opts = vars(args)

    zk_client = KazooClient(hosts=args.zookeeper)
    zk_client.start()

    endpoint = spb.EndpointDef(service_host=args.service_host,service_port=args.service_port)
    predictor = spb.PredictorDef(endpoint=endpoint)
    dep = spb.DeploymentDef(oauth_key=args.oauth_key,oauth_secret=args.oauth_secret,predictor=predictor,id=args.deployment_id,name="test deployment")
    json_string = json_format.MessageToJson(dep)

    node_set(zk_client,"/deployments/"+args.deployment_id,json_string)


