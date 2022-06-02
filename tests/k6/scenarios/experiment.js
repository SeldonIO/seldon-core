import {connectScheduler, disconnectScheduler, loadModel, unloadModel, loadExperiment, unloadExperiment} from '../components/scheduler.js'
import {inferHttpLoop, inferGrpcLoop} from '../components/v2.js'
import {getConfig} from '../components/settings.js'
import {generateExperiment } from '../components/model.js'
import { vu, scenario } from 'k6/execution';

export function setup() {
    return getConfig()
}

export default function (config) {
    const experimentName = config.experimentNamePrefix + scenario.iterationInTest.toString()
    const modelName1 = config.modelNamePrefix + scenario.iterationInTest.toString() + "_1"
    const modelName2 = config.modelNamePrefix + scenario.iterationInTest.toString() + "_2"
    const experiment = generateExperiment(experimentName, config.modelType, modelName1, modelName2, vu.idInTest, 1,
        config.isSchedulerProxy, config.modelMemoryBytes, config.inferBatchSize)
    const model1Defn = experiment.model1Defn
    const model2Defn = experiment.model2Defn
    const experimentDefn = experiment.experimentDefn
    const schedulerEndpoint = config.schedulerEndpoint
    const httpEndpoint = config.inferHttpEndpoint
    const grpcEndpoint = config.inferGrpcEndpoint

    if (config.loadExperiment || config.unloadExperiment) {
        connectScheduler(schedulerEndpoint)
    }

    if (config.loadExperiment) {
        loadModel(modelName1, model1Defn, true)
        loadModel(modelName2, model2Defn, true)
        loadExperiment(experimentName, experimentDefn, true)
    }

    if (config.infer) {
        inferHttpLoop(httpEndpoint, modelName1, experiment.inference.http, config.inferHttpIterations)
        inferGrpcLoop(grpcEndpoint, modelName1, experiment.inference.grpc, config.inferGrpcIterations)
    }

    if (config.unloadExperiment) {
        unloadModel(modelName1, true)
        unloadModel(modelName2, true)
        unloadExperiment(experimentName, true)
    }

    if (config.loadExperiment || config.unloadExperiment) {
        disconnectScheduler()
    }
}