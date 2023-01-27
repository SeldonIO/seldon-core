import {connectScheduler, disconnectScheduler, loadModel, unloadModel} from '../components/scheduler.js'
import {inferHttpLoop, inferGrpcLoop} from '../components/v2.js'
import {getConfig} from '../components/settings.js'
import {generateModel } from '../components/model.js'
import { vu, scenario } from 'k6/execution';

export function setup() {
    return getConfig()
}

export default function (config) {
    // only assume one model type in this scenario
    const idx = 0
    const modelName = config.modelNamePrefix[idx] + scenario.iterationInTest.toString()
    const model = generateModel(config.modelType[idx], modelName, vu.idInTest, 1, 
        config.isSchedulerProxy, config.modelMemoryBytes[idx], config.inferBatchSize[idx])
    const modelDefn = model.modelDefn
    const schedulerEndpoint = config.schedulerEndpoint
    const httpEndpoint = config.inferHttpEndpoint
    const grpcEndpoint = config.inferGrpcEndpoint

    if (config.loadModel || config.unloadModel) {
        connectScheduler(schedulerEndpoint)
    }

    if (config.loadModel) {
        loadModel(modelName, modelDefn, true)
    }

    if (config.infer) {
        inferHttpLoop(httpEndpoint, modelName, model.inference.http, config.inferHttpIterations)
        inferGrpcLoop(grpcEndpoint, modelName, model.inference.grpc, config.inferGrpcIterations)
    }

    if (config.unloadModel) {
        unloadModel(modelName, true)
    }

    if (config.loadModel || config.unloadModel) {
        disconnectScheduler()
    }
}