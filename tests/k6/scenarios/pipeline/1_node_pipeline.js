import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../components/k8s.js';
import * as scheduler from '../../components/scheduler.js'

import { getConfig } from '../../components/settings.js'
import {connectControlPlaneOps,
} from '../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../components/model.js';
import {inferHttp, setupK6, tearDownK6} from "../../components/v2.js";
import {generateServer} from "../../components/k8s.js";

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
    iterations: 50000,
}

const modelType = 'echo'
const modelName = 'tests-pipeline-1-node-echo';
const pipelineName = 'tests-pipeline-1-node-echo-pipeline';
const serverName = "autotest-mlserver";
const modelServerReplicas = 5;

export function setup() {
    return setupK6(function (config) {
        k8s.init()

        const ctl = connectControlPlaneOps(config)

        const modelParams = [
            {
                name: 'response_length',
                value: (10).toString(),
            }
        ]

        const server = generateServer(serverName, "mlserver", modelServerReplicas, 1, modelServerReplicas)
        ctl.unloadServerFn(server.object.metadata.name, true, true)
        ctl.loadServerFn(server.yaml, server.object.metadata.name, true, true, 30)

        const pipeline = generateMultiModelPipelineYaml(1, modelType, pipelineName,
            modelName, modelParams, config.modelName, modelServerReplicas, serverName)


        pipeline.modelCRYaml.forEach(model => {
            ctl.unloadModelFn(model.metadata.name, true)
            ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
        })

        ctl.unloadPipelineFn(pipeline.pipelineName, true)
        ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)
        return config
    }, {
        "useKubeControlPlane": true,
    })
}

export default function (config) {
    const inferPayload1 = getModelInferencePayload(modelType, 1)
    inferHttp(config.inferHttpEndpoint, pipelineName, inferPayload1.http, true, true, config.debug, config.requestIDPrefix)
}

export function teardown(config) {
    tearDownK6(config, function (config) {
        const ctl = connectControlPlaneOps(config)
        ctl.unloadServerFn(serverName, true, false)
        ctl.unloadModelFn(modelName, true)
        ctl.unloadPipelineFn(pipelineName, false)
    })
}
