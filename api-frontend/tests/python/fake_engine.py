import os
import sys, getopt, argparse
import logging
from random import randint,random
import json
import prediction_pb2 as ppb
from google.protobuf import json_format
from flask import Flask
from flask import request

app = Flask(__name__)

@app.route('/api/v0.1/predictions',methods=['GET','POST'])
def hello_world():
    if not request.args.get('json') is None:
        jStr = request.args.get('json')
    else:
        jStr = request.data
    print jStr
    j = json.loads(jStr)

    meta = ppb.PredictionResponseMetaDef(puid="puid-1")
    response = ppb.PredictionResponseDef(meta=meta)
    json_string = json_format.MessageToJson(response)
    print json_string

    return json_string

if __name__ == '__main__':
    import logging
    logger = logging.getLogger()
    logging.basicConfig(format='%(asctime)s : %(levelname)s : %(name)s : %(message)s', level=logging.DEBUG)
    logger.setLevel(logging.INFO)

    parser = argparse.ArgumentParser(prog='create_replay')
    parser.add_argument('--engine-port', help='service port', type=int, default=5000)

    args = parser.parse_args()
    opts = vars(args)

    app.run(host="0.0.0.0", port=args.engine_port,debug=True)



