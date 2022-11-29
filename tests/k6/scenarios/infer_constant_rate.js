import { getConfig } from '../components/settings.js'
import { doInfer, setupBase, teardownBase, getVersionSuffix } from '../components/utils.js'

export const options = {
    thresholds: {
        'http_req_duration{scenario:default}': [`max>=0`],
        'http_reqs{scenario:default}': [],
        'grpc_req_duration{scenario:default}': [`max>=0`],
        'data_received{scenario:default}': [],
        'data_sent{scenario:default}': [],
    },
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: getConfig().requestRate,
            timeUnit: '1s',
            duration: getConfig().constantRateDurationSeconds.toString()+'s',
            preAllocatedVUs: 1, // how large the initial pool of VUs would be
            maxVUs: 1000, // if the preAllocatedVUs are not enough, we can initialize more
        },
    },
};

export function setup() {
    const config = getConfig()

    setupBase(config)

    return config
}

export default function (config) {
    const modelId = Math.floor(Math.random() * config.maxNumModels)
    const modelName = config.modelNamePrefix + modelId.toString()

    const modelNameWithVersion = modelName + getVersionSuffix(config)  // first version

    if (config.inferType === "REST") {
        doInfer(modelName, modelNameWithVersion, config, true)
    } else {
        doInfer(modelName, modelNameWithVersion, config, false)
    }
}

export function teardown(config) {
    teardownBase(config)
}
