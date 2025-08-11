import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../../components/k8s.js';
import * as scheduler from '../../../components/scheduler.js'

import { getConfig } from '../../../components/settings.js'
import {connectControlPlaneOps,
} from '../../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../../components/model.js';
import {inferHttp} from "../../../components/v2.js";

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
    iterations: 10000,
}

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

    const pipeline1 = generateMultiModelPipelineYaml(5, "echo", "echo1", modelParams, config.modelName, 1, 1)


    pipeline1.modelCRYaml.forEach(model => {
        ctl.unloadModelFn(model.metadata.name, true)
        ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
    })

    ctl.unloadPipelineFn(pipeline1.pipelineName, true)
    ctl.loadPipelineFn(pipeline1.pipelineName, pipeline1.pipelineCRYaml, true, true)

    const pipeline2 = generateMultiModelPipelineYaml(5, "echo", "echo2", modelParams, config.modelName, 1, 1)


    pipeline2.modelCRYaml.forEach(model => {
        ctl.unloadModelFn(model.metadata.name, true)
        ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
    })

    ctl.unloadPipelineFn(pipeline2.pipelineName, true)
    ctl.loadPipelineFn(pipeline2.pipelineName, pipeline2.pipelineCRYaml, true, true)

    return config
}

export default function (config) {
    if (Math.random() > 0.5) {
        const inferPayload1 = getModelInferencePayload("echo", 1)
        inferHttp(config.inferHttpEndpoint, "echo1", inferPayload1.http, true, 'pipeline', config.debug)
        return
    }

    const inferPayload2 = getModelInferencePayload("echo", 1)
    inferHttp(config.inferHttpEndpoint, "echo2", inferPayload2.http, true, 'pipeline', config.debug)
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
