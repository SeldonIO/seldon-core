import * as k8s from '../components/k8s.js';
import * as scheduler from '../components/scheduler.js'

import { getConfig } from '../components/settings.js'
import {connectControlPlaneOps,
} from '../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../components/model.js';
import {inferHttp} from "../components/v2.js";

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
    //duration: '5s',
    teardownTimeout: '6000s',
    iterations: 200000,
}

export function setup() {
    k8s.init()
    var config = getConfig()

    const ctl = connectControlPlaneOps(config)

    const modelParams = [
        {
            name: 'response_length',
            value: '50',
        }
    ]

    const pipeline = generateMultiModelPipelineYaml(5, "echo", "echo", modelParams, config.modelName, 1, 1)


    // pipeline.modelCRYaml.forEach(model => {
    //     ctl.unloadModelFn(model.metadata.name, true)
    //     ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
    // })
    //
    // ctl.unloadPipelineFn(pipeline.pipelineName, true)
    //ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)
    config.pipelineName = pipeline.pipelineName
    return config
}

export default function (config) {
    const inferPayload = getModelInferencePayload("echo", 1)
    inferHttp(config.inferHttpEndpoint, "echo", inferPayload.http, true, 'pipeline', config.debug)
}

export function teardown(config) {
    const ctl = connectControlPlaneOps(config)
    //
    // let modelNames = k8s.getExistingModelNames(config.modelName)
    // modelNames.forEach(modelName => {
    //    ctl.unloadModelFn(modelName, false)
    // })
    //
    // let pipelineNames = k8s.getExistingPipelineNames(pipelineName)
    // pipelineNames.forEach(pipelineName => {
    //     ctl.unloadPipelineFn(pipelineName, false)
    // })

    scheduler.disconnectScheduler()
}
