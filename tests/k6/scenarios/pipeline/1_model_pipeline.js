import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../components/k8s.js';
import {connectControlPlaneOps,
} from '../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../components/model.js';
import {connectV2Grpc, inferGrpc, inferHttp, setupK6, tearDownK6} from "../../components/v2.js";
import {generateServer} from "../../components/k8s.js";
import { sleep } from 'k6'

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
    iterations: 50000,
}

const modelType = 'echo'
const modelName = 'alpha-model';
const pipelineName = 'alpha-pipeline';
const serverName = "alpha-mlserver";

// A 1 model pipeline, model is a simple echo model
// which returns a constant string in the response.
//
// Environment variables config options:
// - INFERENCE_SERVER_REPLICAS
// - INFERENCE_SERVER_MIN_REPLICAS
// - INFERENCE_SERVER_MAX_REPLICAS
// - MODEL_REPLICAS
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

        const server = generateServer(serverName, "mlserver",  config.replicas.inferenceServer.actual,
            config.replicas.inferenceServer.min, config.replicas.inferenceServer.max)
        ctl.unloadServerFn(server.object.metadata.name, true, true)
        ctl.loadServerFn(server.yaml, server.object.metadata.name, true, true, 45)

        const pipeline = generateMultiModelPipelineYaml(1, modelType, pipelineName,
            modelName, modelParams, config.modelName, config.replicas.model, serverName)


        pipeline.modelCRYaml.forEach(model => {
            console.log("deleting model")
            ctl.unloadModelFn(model.metadata.name, true)
            console.log("deleted model")
            ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
            console.log("created model")
        })

        console.log("deleting pipeline")
        ctl.unloadPipelineFn(pipeline.pipelineName, true)
        console.log("creating pipeline")
        ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)

        return config
    }, {
        "useKubeControlPlane": true,
    })
}

let gRPCConnected = false

export default function (config) {
    const inferPayload1 = getModelInferencePayload(modelType, 1)

    if (config.useGRPC) {
        if (!gRPCConnected) {
            connectV2Grpc(config.inferGrpcEndpoint)
            gRPCConnected = true
        }
        inferGrpc(pipelineName, inferPayload1.grpc, true, true)
        return
    }
    inferHttp(config.inferHttpEndpoint, pipelineName, inferPayload1.http, true, true, config.debug, config.requestIDPrefix)
}

export function teardown(config) {
    tearDownK6(config, function (config) {
        const ctl = connectControlPlaneOps(config)
        ctl.unloadServerFn(serverName, true, false)
        ctl.unloadModelFn(modelName + "-1", true)
        ctl.unloadPipelineFn(pipelineName, false)
    })
}
