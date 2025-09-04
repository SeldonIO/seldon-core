// This test does following:
//
// - model gw replicas => 2
// - dataflow engine replicas => 2
// - scales pipeline gw replicas => 2
// - scales server replicas => 1
// - sets up 2 pipeline of 5 nodes
// - sends 5000 requests, split across both pipelines

import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../../components/k8s.js';
import {connectControlPlaneOps,
} from '../../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../../components/model.js';
import {inferHttp, setupK6, tearDownK6} from "../../../components/v2.js";
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
    // TODO put back to 5000
    iterations: 500,
}

const modelNamePrefix1 = 'automatedtests-pipeline-1-node-echo'
const modelNamePrefix2 = 'automatedtests-pipeline-2-node-echo'
const pipelineName1 = modelNamePrefix1 + '-pipeline';
const pipelineName2 = modelNamePrefix2 + '-pipeline';
const serverName = "autotest-mlserver"
const modelType = 'echo'


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


        const pipeline1 = generateMultiModelPipelineYaml(5, modelType, pipelineName1,modelNamePrefix1, modelParams, config.modelName, 1, serverName)
        pipeline1.modelCRYaml.forEach(model => {
            ctl.unloadModelFn(model.metadata.name, true)
            ctl.loadModelFn(model.metadata.name, yamlDump(model), false, false)
        })

        ctl.unloadPipelineFn(pipeline1.pipelineName, true)
        ctl.loadPipelineFn(pipeline1.pipelineName, pipeline1.pipelineCRYaml, false, false)

        const pipeline2 = generateMultiModelPipelineYaml(5, modelType, pipelineName2,modelNamePrefix2, modelParams, config.modelName, 1, serverName)

        pipeline2.modelCRYaml.forEach(model => {
            ctl.unloadModelFn(model.metadata.name, true)
            ctl.loadModelFn(model.metadata.name, yamlDump(model), false, false)
        })

        const server = generateServer(serverName, "mlserver", 1, 1, 1)
        ctl.unloadServerFn(server.object.metadata.name, true, true)
        ctl.loadServerFn(server.yaml, server.object.metadata.name, true, true, 35)

        ctl.unloadPipelineFn(pipeline2.pipelineName, true)
        ctl.loadPipelineFn(pipeline2.pipelineName, pipeline2.pipelineCRYaml, true, true)


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
    if (Math.random() > 0.5) {
        const inferPayload1 = getModelInferencePayload(modelType, 1)
        inferHttp(config.inferHttpEndpoint, modelNamePrefix1, inferPayload1.http, true, true, config.debug)
        return
    }

    const inferPayload2 = getModelInferencePayload(modelType, 1)
    inferHttp(config.inferHttpEndpoint, modelNamePrefix2, inferPayload2.http, true, true, config.debug)
}

export function teardown(config) {
    tearDownK6(config, function (config) {
        const ctl = connectControlPlaneOps(config)

        ctl.unloadServerFn(serverName, false, false)

        for (const modelName of [modelNamePrefix1, modelNamePrefix2]) {
            let modelNames = k8s.getExistingModelNames(modelName)
            modelNames.forEach(modelName => {
                console.log("deleting model ", modelName)
                ctl.unloadModelFn(modelName, false)
            })
        }

        for (const pipelineName of [pipelineName1, pipelineName2]) {
            let pipelineNames = k8s.getExistingPipelineNames(pipelineName)
            pipelineNames.forEach(pipelineName => {
                console.log("deleting pipeline ", pipelineName)
                ctl.unloadPipelineFn(pipelineName, false)
            })
        }
    })
}
