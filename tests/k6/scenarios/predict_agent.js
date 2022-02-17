import {connectScheduler, disconnectScheduler, loadModel, unloadModel} from '../components/scheduler_proxy.js'
import {inferHttpLoop, inferGrpcLoop, modelStatusHttp} from '../components/v2.js'
import {getConfig} from '../components/settings.js'
import {generateModel} from '../components/model.js'
import {sleep} from 'k6';

// workaround: https://community.k6.io/t/exclude-http-requests-made-in-the-setup-and-teardown-functions/1525
export let options = {
    thresholds: {
        'http_req_duration{scenario:default}': [`max>=0`],
        'grpc_req_duration{scenario:default}': [`max>=0`],
    },
    setupTimeout: '600s',
    duration: '30m'
}

export function setup() {
    const config = getConfig()

    if (config.loadModel) {
        connectScheduler(config.schedulerEndpoint)

        for (let i = 0; i < config.maxNumModels; i++) {
            const modelName = "model" + i.toString()
            const model = generateModel(config.modelType, modelName, 1, 1, true)
            const modelDefn = model.modelDefn

            loadModel(modelName, modelDefn, false)
            sleep(1)
        }

        // warm up
        for (let i = 0; i < config.maxNumModels; i++) {
            const modelName = "model" + i.toString() + "_0"  // first version
            const model = generateModel(config.modelType, modelName, 1, 1, true)
            const modelDefn = model.modelDefn


            while (modelStatusHttp(config.inferHttpEndpoint, modelName, false) !== 200) {
                sleep(1)
            }

            inferHttpLoop(config.inferHttpEndpoint, modelName, model.inference.http, 1, false)
        }

        disconnectScheduler()
    }

    return config
}

export default function (config) {
    const modelId = Math.floor(Math.random() * config.maxNumModels)
    const modelName = "model" + modelId.toString() + "_0"  // first version
    const model = generateModel(config.modelType, modelName, 1, 1, true)
    const httpEndpoint = config.inferHttpEndpoint
    const grpcEndpoint = config.inferGrpcEndpoint

    if (config.infer) {
        inferHttpLoop(httpEndpoint, modelName, model.inference.http, config.inferHttpIterations, false)
        inferGrpcLoop(grpcEndpoint, modelName, model.inference.grpc, config.inferGrpcIterations, false)
    }
  
}

export function teardown(config) {
    if (config.unloadModel) {
        connectScheduler(config.schedulerEndpoint)

        for (let i = 0; i < config.maxNumModels; i++) {
            const modelName = "model" + i.toString()
            unloadModel(modelName, true)
        }

        disconnectScheduler()
    }
  }