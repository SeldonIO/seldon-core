// This test does following:
//
// - scales model gw replicas => 2
// - scales dataflow engine replicas => 2
// - scales pipeline gw replicas => 2
// - scales server replicas => 5
// - sets up pipeline of 1 node
// - sends 5000 requests

import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../../components/k8s.js';
import {connectControlPlaneOps,
} from '../../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../../components/model.js';
import {inferHttp, setupK6} from "../../../components/v2.js";
import {generateSeldonRuntime, generateServer} from "../../../components/k8s.js";
import { sleep } from 'k6';

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
    teardownTimeout: '6000s',
    iterations: 5000,
}

const modelType = 'echo'
const modelName = 'tests-pipeline-1-node-echo';
const pipelineName = 'tests-pipeline-1-node-echo-pipeline';
const serverName = "autotest-mlserver"

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

        const modelServerReplicas = 5


        const pipeline = generateMultiModelPipelineYaml(1, modelType, pipelineName, modelName, modelParams, config.modelName, modelServerReplicas, serverName)
        pipeline.modelCRYaml.forEach(model => {
            // we MUST unload/load models before loading server, otherwise it causes Server to scale down to 1, bug or design?
            ctl.unloadModelFn(model.metadata.name, true)
            ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
        })

        const server = generateServer(serverName, "mlserver", modelServerReplicas, 1, modelServerReplicas)
        ctl.unloadServerFn(server.object.metadata.name, true, true)
        ctl.loadServerFn(server.yaml, server.object.metadata.name, true, true, 30)

        ctl.unloadPipelineFn(pipeline.pipelineName, true)
        ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)
        // TODO we have to wait, as server's eagerly load models before they have IP available,
        //   waiting for this fix to be merged then can take out sleep https://github.com/SeldonIO/seldon-core/pull/6636
        sleep(10)


        const replicaModelGw = 2;
        const replicaDataFlowEngine = 2;
        const replicaPipeLineGw = 2;

        // we have to scale the model-gw, dataflow-engine, pipeline-gw AFTER we have deployed the pipelines
        // as otherwise the seldon controller will prohibit the scaling and default to 1 replica as there's no
        // pipelines deployed
        const seldonRuneTime = generateSeldonRuntime(replicaModelGw,replicaPipeLineGw,replicaDataFlowEngine)
        ctl.loadSeldonRuntimeFn(seldonRuneTime.object, true, true)

        return config
    }, {
        "useKubeControlPlane" : true,
    })
}

export default function (config) {
    const inferPayload1 = getModelInferencePayload(modelType, 1)
    inferHttp(config.inferHttpEndpoint, modelName, inferPayload1.http, true, true, config.debug, config.requestIDPrefix)
}

export function teardown(config) {
    const ctl = connectControlPlaneOps(config)

    ctl.unloadServerFn(serverName, false, false)

    let modelNames = k8s.getExistingModelNames(modelName)
    modelNames.forEach(modelName => {
       ctl.unloadModelFn(modelName, false)
    })

    let pipelineNames = k8s.getExistingPipelineNames(pipelineName)
    pipelineNames.forEach(pipelineName => {
        ctl.unloadPipelineFn(pipelineName, false)
    })
}
