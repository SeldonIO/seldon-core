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

    for (let i = 0; i < config.maxNumModels; i++) {
        var modelName = config.modelNamePrefix + i.toString()

        var modelNameWithVersion = modelName + getVersionSuffix(config)  // first version
        doInfer(modelName, modelNameWithVersion, config, false)
    }
}

export function teardown(config) {
    teardownBase(config)
  }