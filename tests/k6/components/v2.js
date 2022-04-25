import http from 'k6/http';
import { check } from 'k6';
import grpc from 'k6/net/grpc';
import { generatePipelineName } from '../components/model.js';

const v2Client = new grpc.Client();
v2Client.load(['../../../apis/mlops/v2_dataplane/'], 'v2_dataplane.proto');

export function inferHttp(endpoint, modelName, payload, viaEnvoy, pipelineSuffix) {
    const url = endpoint + "/v2/models/"+modelName+"/infer"
    const payloadStr = JSON.stringify(payload);
    var headers = {
        'Content-Type': 'application/json',
        'Host': modelName,
        // we add here either .model or .pipeline to test dataflow
        'seldon-model': generateDataFlowName(modelName, pipelineSuffix),
    };
    if (viaEnvoy != true) {
        headers['seldon-internal-model'] = modelName
    }
    const params = {
        headers:  headers
    };
    //console.log("URL:",url,"Payload:",payloadStr,"Params:",JSON.stringify(params))
    const response = http.post(url, payloadStr, params);
    //console.log(response.status, JSON.stringify(response.error), response.body)
    check(response, {'model http prediction success': (r) => r.status === 200});
}

export function inferHttpLoop(endpoint, modelName, payload, iterations, viaEnvoy = true, pipelineSuffix = "") {
    for (let i = 0; i < iterations; i++) {
        inferHttp(endpoint, modelName, payload, viaEnvoy, pipelineSuffix)
    }
}

export function inferGrpc(modelName, payload, viaEnvoy, pipelineSuffix) {
    var headers = {
        // we add here either .model or .pipeline to test dataflow
        'seldon-model': generateDataFlowName(modelName, pipelineSuffix),
    };
    if (viaEnvoy != true) {
        headers['seldon-internal-model'] = modelName
    }
    const params = {
        headers: headers
    };
    payload.model_name = modelName
    const response = v2Client.invoke('inference.GRPCInferenceService/ModelInfer', payload, params);
    //console.log(response.status, JSON.stringify(response.error),  JSON.stringify(response.message))
    check(response, {'model grpc prediction success': (r) => r && r.status === grpc.StatusOK})
}

export function inferGrpcLoop(endpoint, modelName, payload, iterations, viaEnvoy = true, pipelineSuffix = "") {
    connectV2Grpc(endpoint)
    for (let i = 0; i < iterations; i++) {
        inferGrpc(modelName, payload, viaEnvoy, pipelineSuffix)
    }
    disconnectV2Grpc()
}

export function modelStatusHttp(endpoint, modelName, viaEnvoy = true) {
    const url = endpoint + "/v2/models/"+modelName+"/ready"
    var headers = {
        'Content-Type': 'application/json',
        'Host' : modelName,
        'seldon-model' : modelName,
    };
    if (viaEnvoy != true) {
        headers['seldon-internal-model'] = modelName
    }
    const params = {
        headers: headers
    };
    const response = http.get(url, params);
    return response.status
}

export function connectV2Grpc(endpoint) {
    v2Client.connect(endpoint, {
        plaintext: true
    });
    return v2Client
}

export function disconnectV2Grpc() {
    v2Client.close();
}

export function generateDataFlowName(modelName, suffix) {
    if (suffix=="") {
        return modelName
    } else {
        if (suffix=="pipeline") {
            // we add -pipeline to the model name in the case of constructing a pipeline
            return generatePipelineName(modelName)+"."+suffix
        } else {
            // in this dataflow case we access the model directly
            return modelName+"."+suffix
        }
    }
}

