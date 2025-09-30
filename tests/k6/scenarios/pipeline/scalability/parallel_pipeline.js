import {dump as yamlDump} from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../../components/k8s.js';
import {
    connectControlPlaneOps,
} from '../../../components/utils.js'
import {
    generateMultiModelParallelPipelineYaml,
    getModelInferencePayload
} from '../../../components/model.js';
import {inferHttp, setupK6} from "../../../components/v2.js";
import {
    awaitPipelineStatus,
    generateSeldonConfig,
    generateSeldonRuntime,
    generateServer
} from "../../../components/k8s.js";

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

const inputModelName1 = 'hotel-model-1'
const inputModelName2 = 'hotel-model-2'
const inputModelType = 'synth'
const outputModelType = 'synth'
const outputModelName = 'hotel-model-combiner'
const pipelineName = 'hotel-pipeline';
const serverName = "hotel-mlserver"


// Sets up parallel pipeline:
// 2 input models runs in parallel
// and we wait for both outputs to input
// into the final model. Each model has an average
// configured latency of 0.1 seconds with a standard
// deviation of 1.
export function setup() {
    return setupK6(function (config) {
        k8s.init()

        const ctl = connectControlPlaneOps(config)

        const modelParams = [
            {
                name: 'predict_latency_dist',
                value: 'normal'
            },
            {
                name: 'work_type',
                value: 'async_cpu_busy_iter',
            },
            {
                name: 'predict_latency_avg_us',
                value: '100000'
            },
            {
                name: "predict_latency_sd_us",
                value: "1"
            }
        ]

        const server = generateServer(serverName, "mlserver", config.replicas.inferenceServer.actual,
            config.replicas.inferenceServer.min, config.replicas.inferenceServer.max)

        console.log("Deleting server")
        ctl.unloadServerFn(server.object.metadata.name, true, true)
        console.log("Deleted server")
        console.log("Creating server")
        ctl.loadServerFn(server.yaml, server.object.metadata.name, true, true, 30)
        console.log("Server created")

        const pipeline = generateMultiModelParallelPipelineYaml(pipelineName,
            [inputModelName1, inputModelName2], inputModelType, modelParams, outputModelName,
            outputModelType, 1, config.replicas.model, serverName)

        pipeline.modelCRYaml.forEach(model => {
            ctl.unloadModelFn(model.metadata.name, true)
            ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
        })

        ctl.unloadPipelineFn(pipeline.pipelineName, true)
        ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)


        const replicaModelGw = config.replicas.modelGw
        const replicaDataFlowEngine = config.replicas.dataFlowEngine
        const replicaPipeLineGw = config.replicas.pipelineGw

        const seldonConfig = generateSeldonConfig(config.dataflow.limits.memory, config.modelGateway.numOfWorkers)
        ctl.loadSeldonConfigFn(seldonConfig.object, true)


        // we have to scale the model-gw, dataflow-engine, pipeline-gw AFTER we have deployed the pipelines
        // as otherwise the seldon controller will prohibit the scaling and default to 1 replica as there's no
        // pipelines deployed
        const seldonRuneTime = generateSeldonRuntime(replicaModelGw, replicaPipeLineGw, replicaDataFlowEngine, "4G")
        ctl.loadSeldonRuntimeFn(seldonRuneTime.object, true, true)

        // wait for pipeline to be ready since scaling data-plane services
        awaitPipelineStatus(pipelineName, "Ready")

        return config
    }, {
        "useKubeControlPlane": true,
    })
}

export default function (config) {
    const inferPayload = getModelInferencePayload(inputModelType, 1)
    inferHttp(config.inferHttpEndpoint, pipelineName, inferPayload.http, true, true, config.debug)
}

export function teardown(config) {
    const ctl = connectControlPlaneOps(config)

    for (const modelName of [inputModelName1, inputModelName2, outputModelName]) {
        console.log("deleting model ", modelName)
        ctl.unloadModelFn(modelName, true)
    }

    console.log("deleting pipeline ", pipelineName)
    ctl.unloadPipelineFn(pipelineName, false)

    ctl.unloadServerFn(serverName, false, false)
}
