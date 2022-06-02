import {connectScheduler, disconnectScheduler, loadModel, unloadModel, loadPipeline, unloadPipeline} from '../components/scheduler.js'
import {inferHttpLoop, inferGrpcLoop} from '../components/v2.js'
import {getConfig} from '../components/settings.js'
import {generateModel, generatePipelineName} from '../components/model.js'
import { vu, scenario } from 'k6/execution';

export function setup() {
    return getConfig()
}

export default function (config) {
    const modelName = config.modelNamePrefix + scenario.iterationInTest.toString()
    const model = generateModel(config.modelType, modelName, vu.idInTest, 1,
        config.isSchedulerProxy, config.modelMemoryBytes, config.inferBatchSize)
    const modelDefn = model.modelDefn
    const pipelineDefn = model.pipelineDefn
    const schedulerEndpoint = config.schedulerEndpoint
    const httpEndpoint = config.inferHttpEndpoint
    const grpcEndpoint = config.inferGrpcEndpoint

    if (config.loadModel || config.unloadModel) {
        connectScheduler(schedulerEndpoint)
    }

    if (config.loadModel) {
        loadModel(modelName, modelDefn, true)
        loadPipeline(generatePipelineName(modelName), pipelineDefn, true)
    }

    if (config.infer) {
        inferHttpLoop(httpEndpoint, modelName, model.inference.http, config.inferHttpIterations, config.isEnvoy, "pipeline")
        inferGrpcLoop(grpcEndpoint, modelName, model.inference.grpc, config.inferGrpcIterations, config.isEnvoy, "pipeline")
    }

    if (config.unloadModel) {
        unloadModel(modelName, true)
        unloadPipeline(generatePipelineName(modelName), true)
    }

    if (config.loadModel || config.unloadModel) {
        disconnectScheduler()
    }
}