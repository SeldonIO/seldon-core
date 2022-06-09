import {inferHttpLoop, inferGrpcLoop, inferHttp, inferGrpc, connectV2Grpc, disconnectV2Grpc} from '../components/v2.js'
import {getConfig} from '../components/settings.js'
import {generateModel } from '../components/model.js'
import { vu, scenario } from 'k6/execution';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 10,
            timeUnit: '1s',
            duration: '5s',
            preAllocatedVUs: 1, // how large the initial pool of VUs would be
            maxVUs: 100, // if the preAllocatedVUs are not enough, we can initialize more
        },
    },
};

export function setup() {
    return getConfig()
}

export default function (config) {
    const modelIdx = randomIntBetween(config.modelStartIdx, config.modelEndIdx)
    const modelName = config.modelNamePrefix + modelIdx.toString()
    const model = generateModel(config.modelType, modelName, 0, 1,
        config.isSchedulerProxy, config.modelMemoryBytes, config.inferBatchSize)
    const modelDefn = model.modelDefn
    const httpEndpoint = config.inferHttpEndpoint
    const grpcEndpoint = config.inferGrpcEndpoint

    if (config.inferType === "REST") {
        if (config.modelName !== "") {
            inferHttp(httpEndpoint, config.modelName, model.inference.http, true, "")
        } else {
            inferHttp(httpEndpoint, modelName, model.inference.http, true, "")
        }
    } else {
        connectV2Grpc(grpcEndpoint)
        if (config.modelName !== "") {
            inferGrpc(config.modelName, model.inference.grpc, true, "")
        } else {
            inferGrpc(modelName, model.inference.grpc, true, "")
        }
        disconnectV2Grpc()
    }
}