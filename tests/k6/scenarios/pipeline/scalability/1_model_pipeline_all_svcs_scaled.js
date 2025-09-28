import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../../components/k8s.js';
import {connectControlPlaneOps,
} from '../../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../../components/model.js';
import {inferHttp, setupK6, tearDownK6} from "../../../components/v2.js";
import {awaitPipelineStatus, generateSeldonRuntime, generateServer} from "../../../components/k8s.js";

// workaround: https://community.k6.io/t/exclude-http-requests-made-in-the-setup-and-teardown-functions/1525
export let options = {
    thresholds: {
        'http_req_duration{scenario:default}': [`max>=0`],
        'http_reqs{scenario:default}': [],
        'grpc_req_duration{scenario:default}': [`max>=0`],
        'data_received{scenario:default}': [],
        'data_sent{scenario:default}': [],
    },
    setupTimeout: '120s',
    teardownTimeout: '120s',
    iterations: 5000,
}

const modelType = 'echo'
const modelName = 'delta-model';
const pipelineName = 'delta-pipeline';
const serverName = "delta-mlserver"

// A 1 model pipeline, model is a simple echo model
// which returns a constant string in the response.
//
// - scales model gw replicas => 2
// - scales dataflow engine replicas => 2
// - scales pipeline gw replicas => 2
// - scales server replicas => 5
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

        const server = generateServer(serverName, "mlserver", modelServerReplicas, modelServerReplicas, modelServerReplicas)
        ctl.unloadServerFn(server.object.metadata.name, true, true)
        ctl.loadServerFn(server.yaml, server.object.metadata.name, true, true, 30)


        const pipeline = generateMultiModelPipelineYaml(1, modelType, pipelineName, modelName, modelParams, config.modelName, modelServerReplicas, serverName)
        pipeline.modelCRYaml.forEach(model => {
            // we MUST unload/load models before loading server, otherwise it causes Server to scale down to 1, bug or design?
            ctl.unloadModelFn(model.metadata.name, true)
            ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
        })

        ctl.unloadPipelineFn(pipeline.pipelineName, true)
        ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)

        const replicaModelGw = 2;
        const replicaDataFlowEngine = 2;
        const replicaPipeLineGw = 2;

        // we have to scale the model-gw, dataflow-engine, pipeline-gw AFTER we have deployed the pipelines
        // as otherwise the seldon controller will prohibit the scaling and default to 1 replica as there's no
        // pipelines deployed
        const seldonRuneTime = generateSeldonRuntime(replicaModelGw, replicaPipeLineGw, replicaDataFlowEngine)
        ctl.loadSeldonRuntimeFn(seldonRuneTime.object, true, true)

        // wait for pipeline to be ready since scaling data-plane services
        awaitPipelineStatus(pipelineName, "Ready")

        return config
    }, {
        "useKubeControlPlane" : true,
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

        let modelNames = k8s.getExistingModelNames(modelName)
        modelNames.forEach(modelName => {
            ctl.unloadModelFn(modelName, false)
        })

        let pipelineNames = k8s.getExistingPipelineNames(pipelineName)
        pipelineNames.forEach(pipelineName => {
            ctl.unloadPipelineFn(pipelineName, false)
        })
    })
}
