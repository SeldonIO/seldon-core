import * as k8s from '../components/k8s.js';
import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as scheduler from '../components/scheduler.js'

import { getConfig } from '../components/settings.js'
import {connectControlPlaneOps,
} from '../components/utils.js'
import {generateMultiModelPipelineYaml} from '../components/model.js';
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
    duration: '30m',
    teardownTimeout: '6000s',
}

let pipelineName;
let inferData;

export function setup() {
    k8s.init()
    var config = getConfig()

    const ctl = connectControlPlaneOps(config)

    const pipeline = generateMultiModelPipelineYaml(5, config.modelType, config.modelName, 1, 1)
    pipelineName = pipeline.pipelineName;
    inferData = pipeline.inference.inference;

    // pipeline.modelCRYaml.forEach(model => {
    //     console.log(yamlDump(model));
    //     ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
    // })

    ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)
    return config
}

export default function (config) {
    const httpEndpoint = config.inferHttpEndpoint
    inferHttp(httpEndpoint, pipelineName, inferData, false, 'pipeline')
}

export function teardown(config) {
    let modelNames = k8s.getExistingModelNames(config.modelName)
    for (var modelName in modelNames) {
        k8s.unloadModel(modelName, false)
    }


    let pipelineNames = k8s.getExistingPipelineNames(pipelineName)
    for (var pipelineName in pipelineNames) {
        k8s.unloadModel(pipelineName, false)
    }

    scheduler.disconnectScheduler()
}
