import { getConfig } from '../components/settings.js'
import { doInfer, setupBase, teardownBase, getVersionSuffix } from '../components/utils.js'

// workaround: https://community.k6.io/t/exclude-http-requests-made-in-the-setup-and-teardown-functions/1525
export let options = {
    thresholds: {
        'http_req_duration{scenario:default}': [`max>=0`],
        'http_reqs{scenario:default}': [],
        'grpc_req_duration{scenario:default}': [`max>=0`],
        'data_received{scenario:default}': [],
        'data_sent{scenario:default}': [],
    },
    setupTimeout: '6000s',
    duration: '30m',
    teardownTimeout: '6000s',
}

export function setup() {
    const config = getConfig()

    setupBase(config)

    return config
}

export default function (config) {
    const numModelTypes = config.modelType.length
    const idx = Math.floor(Math.random() * numModelTypes)
    const modelId = Math.floor(Math.random() * config.maxNumModels[idx])
    const modelName = config.modelNamePrefix[idx] + modelId.toString()

    const modelNameWithVersion = modelName + getVersionSuffix(config.isSchedulerProxy)  // first version

    // the settings here can be either 0/"0" or 1/"1"
    // so we can use them in boolean context to check if a protocol is enabled for the test
    var rest_enabled = Number(config.inferHttpIterations)
    var grpc_enabled = Number(config.inferGrpcIterations)
    if (rest_enabled && grpc_enabled) {
        // if both protocols are enabled, choose one randomly
        const rand = Math.random()
        if (rand > 0.5) {
            doInfer(modelName, modelNameWithVersion, config, true, idx) // rest
        } else {
            doInfer(modelName, modelNameWithVersion, config, false, idx) // grpc
        }
    } else if (rest_enabled) {
        doInfer(modelName, modelNameWithVersion, config, true, idx)
    } else if (grpc_enabled) {
        doInfer(modelName, modelNameWithVersion, config, false, idx)
    } else {
        throw new Error('Both REST and GRPC protocols are disabled!')
    }
}

export function teardown(config) {
    teardownBase(config)
}