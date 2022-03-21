import { sleep } from 'k6';
import { generateModel } from '../components/model.js';
import { connectScheduler, disconnectScheduler, loadModel, unloadModel } from '../components/scheduler.js';
import {
    connectScheduler as connectSchedulerProxy,
    disconnectScheduler as disconnectSchedulerProxy,
    loadModel as loadModelProxy,
    unloadModel as unloadModelProxy
} from '../components/scheduler_proxy.js';
import { inferGrpcLoop, inferHttpLoop, modelStatusHttp } from '../components/v2.js';

export function setupBase(config ) {
    if (config.loadModel) {
        
        var connectSchedulerFn = connectScheduler
        if (config.isSchedulerProxy) {
            connectSchedulerFn = connectSchedulerProxy
        }
        
        connectSchedulerFn(config.schedulerEndpoint)

        for (let i = 0; i < config.maxNumModels; i++) {
            const modelName = "model" + i.toString()
            const model = generateModel(config.modelType, modelName, 1, 1, config.isSchedulerProxy, config.modelMemoryBytes)
            const modelDefn = model.modelDefn

            var loadModelFn = loadModel
            if (config.isSchedulerProxy) {
                connectSchedulerFn = loadModelProxy
            }

            loadModelFn(modelName, modelDefn, false)
            
            sleep(1)
        }

        // warm up
        for (let i = 0; i < config.maxNumModels; i++) {
            const modelName = "model" + i.toString()
            const modelNameWithVersion = modelName + "_1"  // first version
            
            const model = generateModel(config.modelType, modelNameWithVersion, 1, 1, config.isSchedulerProxy, config.modelMemoryBytes)

            while (modelStatusHttp(config.inferHttpEndpoint, config.isEnvoy?modelName:modelNameWithVersion, config.isEnvoy) !== 200) {
                sleep(1)
            }

            inferHttpLoop(config.inferHttpEndpoint, config.isEnvoy?modelName:modelNameWithVersion, model.inference.http, 1, config.isEnvoy)
        }

        var disconnectSchedulerFn = disconnectScheduler
        if (config.isSchedulerProxy) {
            disconnectSchedulerFn = disconnectSchedulerProxy
        }
        disconnectSchedulerFn()
    }
}

export function teardownBase(config ) {
    if (config.unloadModel) {
        var connectSchedulerFn = connectScheduler
        if (config.isSchedulerProxy) {
            connectSchedulerFn = connectSchedulerProxy
        }
        connectSchedulerFn(config.schedulerEndpoint)

        var unloadModelFn = unloadModel
        if (config.isSchedulerProxy) {
            unloadModelFn = unloadModelProxy
        }

        for (let i = 0; i < config.maxNumModels; i++) {
            const modelName = "model" + i.toString()
            unloadModelFn(modelName, true)
        }

        var disconnectSchedulerFn = disconnectScheduler
        if (config.isSchedulerProxy) {
            disconnectSchedulerFn = disconnectSchedulerProxy
        }
        disconnectSchedulerFn()
    }
}

export function doInfer(modelName, modelNameWithVersion, config, isHttp) {
    const model = generateModel(config.modelType, modelName, 1, 1, config.isSchedulerProxy, config.modelMemoryBytes)
    const httpEndpoint = config.inferHttpEndpoint
    const grpcEndpoint = config.inferGrpcEndpoint

    if (config.infer) {
        if (isHttp) {
            inferHttpLoop(httpEndpoint, config.isEnvoy?modelName:modelNameWithVersion, model.inference.http, config.inferHttpIterations, config.isEnvoy)
        } else {
            inferGrpcLoop(grpcEndpoint, config.isEnvoy?modelName:modelNameWithVersion, model.inference.grpc, config.inferGrpcIterations, config.isEnvoy)
        }
    }
}
