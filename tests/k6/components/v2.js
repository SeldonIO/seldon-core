import http from 'k6/http';
import { check } from 'k6';
import grpc from 'k6/net/grpc';
import { generatePipelineName } from '../components/model.js';
import {getConfig} from "./settings.js";
import {connectControlPlaneOps} from "./utils.js";

const v2Client = new grpc.Client();
v2Client.load([import.meta.resolve('../../../apis/mlops/v2_dataplane/')], 'v2_dataplane.proto');


export function inferHttp(endpoint, modelName, payload, viaEnvoy, isPipeline = false, debug = false, requestIDPrefix = null) {
    const url = endpoint + "/v2/models/"+modelName+"/infer"
    const payloadStr = JSON.stringify(payload);
    var metadata = {
        'Content-Type': 'application/json',
        'Host': modelName,
        // we add here either .model or .pipeline to test dataflow
        'seldon-model': generateDataFlowName(modelName, isPipeline),
        // disable response compression
        'Accept-Encoding': 'entity',
    };

    if (requestIDPrefix !== null) {
        metadata['x-request-id'] = requestIDPrefix + "_" + generateRandomString(15);
    }

    if (viaEnvoy != true) {
        metadata['seldon-internal-model'] = modelName
    }
    const params = {
        headers:  metadata
    };

    const response = http.post(url, payloadStr, params);

    if (debug && (response.status !== 200) ) {
        console.error(new Date().toISOString(), "URL:",url, "Status:",response.status, "Payloads:",JSON.stringify(response.error), response.body)
    }

    check(response, {'model http prediction success': (r) => r.status === 200});
}

function generateRandomString(length = 10) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

export function inferHttpLoop(endpoint, modelName, payload, iterations, viaEnvoy = true, viaPipeline = false) {
    for (let i = 0; i < iterations; i++) {
        inferHttp(endpoint, modelName, payload, viaEnvoy, viaPipeline)
    }
}

export function inferGrpc(modelName, payload, viaEnvoy, isPipeline = false) {
    var metadata = {
        // we add here either .model or .pipeline to test dataflow
        'seldon-model': generateDataFlowName(modelName, isPipeline),
    };
    if (viaEnvoy != true) {
        metadata['seldon-internal-model'] = modelName
    }
    const params = {
        metadata: metadata
    };
    payload.model_name = modelName
    const response = v2Client.invoke('inference.GRPCInferenceService/ModelInfer', payload, params);
    check(response, {'model grpc prediction success': (r) => r && r.status === grpc.StatusOK})

    if (response.status !== grpc.StatusOK) {
        console.log(response.error)
    }
}

export function inferGrpcLoop(endpoint, modelName, payload, iterations, viaEnvoy = true, isPipeline = false) {
    connectV2Grpc(endpoint)
    for (let i = 0; i < iterations; i++) {
        inferGrpc(modelName, payload, viaEnvoy, isPipeline)
    }
    disconnectV2Grpc()
}

export function modelStatusHttp(endpoint, modelName, viaEnvoy = true) {
    const url = endpoint + "/v2/models/"+modelName+"/ready"
    var metadata = {
        'Content-Type': 'application/json',
        'Host' : modelName,
        'seldon-model' : modelName,
    };
    if (viaEnvoy != true) {
        metadata['seldon-internal-model'] = modelName
    }
    const params = {
        metadata: metadata
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

export function generateDataFlowName(modelName, isPipeline) {
    if (!isPipeline) {
        return modelName
    }
    // we add .pipeline to the model name in the case of constructing a pipeline
    return generatePipelineName(modelName)+".pipeline"
}

export function setupK6(setup, forceConfig = null) {
    let config = getConfig()
    if (forceConfig !== null && forceConfig instanceof Object) {
        for (const key in forceConfig) {
            config[key] = forceConfig[key]
        }
    }

    if (config.skipSetup) {
        return config
    }
    return setup(config)
}

export function tearDownK6(config, callbackTearDown) {
    if (config.skipTearDown) {
        return
    }
    return callbackTearDown(config)
}