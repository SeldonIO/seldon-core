import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../components/k8s.js';
import * as scheduler from '../../components/scheduler.js'

import { getConfig } from '../../components/settings.js'
import {connectControlPlaneOps,
} from '../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../components/model.js';
import {inferHttp} from "../../components/v2.js";

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
    iterations: 5000,
}

const modelType = 'echo'
const modelName = 'tests-pipeline-1-node-echo';

export function setup() {
    k8s.init()
    var config = getConfig()

    const ctl = connectControlPlaneOps(config)

    if (config.skipSetup) {
        return config
    }

    const modelParams = [
        {
            name: 'response_length',
            value: (10).toString(),
        }
    ]

    const pipeline = generateMultiModelPipelineYaml(1, modelType, modelName, modelParams, config.modelName, 1, 1)


    pipeline.modelCRYaml.forEach(model => {
        ctl.unloadModelFn(model.metadata.name, true)
        ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
    })

    ctl.unloadPipelineFn(pipeline.pipelineName, true)
    ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)
    return config
}

export default function (config) {
    const inferPayload1 = getModelInferencePayload(modelType, 1)
    inferHttp(config.inferHttpEndpoint, modelName, inferPayload1.http, true, 'pipeline', config.debug)
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
