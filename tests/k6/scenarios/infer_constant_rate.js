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
    setupTimeout: '6000s',
    teardownTimeout: '6000s',
};

export function setup() {
    const config = getConfig()

    setupBase(config)

    return config
}

export default function (config) {
    const numModelTypes = config.modelType.length
    var idx = Math.floor(Math.random() * numModelTypes)
    while (config.maxNumModels[idx] == 0) {
        idx = Math.floor(Math.random() * numModelTypes)
    }
    const modelId = Math.floor(Math.random() * config.maxNumModels[idx])
    const modelName = config.modelNamePrefix[idx] + modelId.toString()

    const modelNameWithVersion = modelName + getVersionSuffix(config.isSchedulerProxy)  // first version

    if (config.inferType === "REST") {
        doInfer(modelName, modelNameWithVersion, config, true, idx)
    } else {
        doInfer(modelName, modelNameWithVersion, config, false, idx)
    }
}

export function teardown(config) {
    teardownBase(config)
}
