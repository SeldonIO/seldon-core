import {getConfig} from '../components/settings.js'
import { setupBase, teardownBase, doInfer, getVersionSuffix } from '../components/utils.js'

// workaround: https://community.k6.io/t/exclude-http-requests-made-in-the-setup-and-teardown-functions/1525
export let options = {
    thresholds: {
        'http_req_duration{scenario:default}': [`max>=0`],
        'grpc_req_duration{scenario:default}': [`max>=0`],
    },
    setupTimeout: '600s',
    duration: '30m',
}

export function setup() {
    const config = getConfig()

    setupBase(config)

    return config
}

export default function (config) {
    // only assume one model type in this scenario
    const idx = 0
    for (let i = 0; i < config.maxNumModels[idx]; i++) {
        var modelName = config.modelNamePrefix[idx] + i.toString()

        var modelNameWithVersion = modelName + getVersionSuffix(config.isSchedulerProxy)  // first version
        doInfer(modelName, modelNameWithVersion, config, false, idx)
    }
}

export function teardown(config) {
    teardownBase(config)
  }