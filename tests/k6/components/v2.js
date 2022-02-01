import http from 'k6/http';
import { check } from 'k6';
import grpc from 'k6/net/grpc';

const v2Client = new grpc.Client();
v2Client.load(['../../../scheduler/v2/'], 'grpc_service.proto');

export function inferHttp(endpoint, modelName, payload) {
    const url = endpoint + "/v2/models/"+modelName+"/infer"
    const payloadStr = JSON.stringify(payload);
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Host' : modelName
        },
    };
    //console.debug("URL:",url,"Payload:",payloadStr,"Params:",JSON.stringify(params))
    const response = http.post(url, payloadStr, params);
    check(response, {'model http prediction success': (r) => r.status === 200});
}

export function inferHttpLoop(endpoint, modelName, payload, iterations) {
    for (let i = 0; i < iterations; i++) {
        inferHttp(endpoint, modelName, payload)
    }
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

export function inferGrpc(modelName, payload) {
    const params = {
        headers: {
            'Seldon' : modelName
        },
    };
    payload.model_name = modelName
    const response = v2Client.invoke('inference.GRPCInferenceService/ModelInfer', payload, params);
    //console.log(response.status,JSON.stringify(response.error),response.message)
    check(response, {'model grpc prediction success': (r) => r && r.status === grpc.StatusOK})
}

export function inferGrpcLoop(endpoint, modelName, payload, iterations) {
    connectV2Grpc(endpoint)
    for (let i = 0; i < iterations; i++) {
        inferGrpc(modelName, payload)
    }
    disconnectV2Grpc()
}

