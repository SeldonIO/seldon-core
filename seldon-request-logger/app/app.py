from flask import Flask, request
import sys
import dict_digger
import json

app = Flask(__name__)

import logging
log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)

@app.route("/", methods=['GET','POST'])
def index():
    try:
        content = request.get_json(force=True)

        requestPart = dict_digger.dig(content,'request')

        if requestPart != None:
            transformDataNdarray(requestPart)

        responsePart = dict_digger.dig(content,'response')

        if responsePart != None:
            transformDataNdarray(responsePart)

        #log formatted json to stdout for fluentd collection
        print(str(json.dumps(content)))
        sys.stdout.flush()

        return str(content)
    except Exception as e:
        print(e, file=sys.stderr)
        return 'Error processing input'

def transformDataNdarray(jsonDict):
    ndarray = dict_digger.dig(jsonDict,'data','ndarray')
    names = dict_digger.dig(jsonDict,'data','names')

    #won't transform features unless we have names
    if names == None:
        return
    if ndarray == None:
        return

    #must be single-dimension set of feature values or batch
    if len(ndarray) >= 3:
        return

    jsonDict['elements'] = {}
    if isinstance(ndarray[0], list):
        #we'll assume batch is first dim - row-primary
        for row in ndarray:
            #we iterate through all features - may later have to add a max number
            for num, name in enumerate(names,start=0):
                jsonDict['elements'][name] = row[num]
    else:
        #must be single dimension
        for num, name in enumerate(names,start=0):
            jsonDict['elements'][name] = ndarray[num]

    return

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)