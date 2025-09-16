import {dump as yamlDump} from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../../components/k8s.js';
import {
    connectControlPlaneOps,
} from '../../../components/utils.js'
import {
    generateMultiModelParallelPipelineYaml,
    getModelInferencePayload
} from '../../../components/model.js';
import {inferHttp, setupK6, tearDownK6} from "../../../components/v2.js";
import {awaitPipelineStatus, generateSeldonRuntime, generateServer} from "../../../components/k8s.js";
import exec from 'k6/execution';

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

const inputModelName1 = 'automatedtests-input-1'
const inputModelName2 = 'automatedtests-input-2'
const inputModelType = 'synth'
const outputModelType = 'synth'
const outputModelName = 'automatedtests-combiner'
const pipelineName = 'automatedtests-parallel';
const serverName = "autotest-mlserver"

const replicaModelGw = 1;
const replicaDataFlowEngine = 1;
const replicaPipeLineGw = 1;


export function setup() {
    return setupK6(function (config) {
        k8s.init()

        const ctl = connectControlPlaneOps(config)

        const inputModelParams = [
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

        const server = generateServer(serverName, "mlserver", 1, 1, 1)
        ctl.unloadServerFn(server.object.metadata.name, true, true)
        ctl.loadServerFn(server.yaml, server.object.metadata.name, true, true, 30)

        const pipeline = generateMultiModelParallelPipelineYaml(pipelineName,
            [inputModelName1, inputModelName2], inputModelType, inputModelParams, outputModelName,
            outputModelType, 1, 1, serverName)

        pipeline.modelCRYaml.forEach(model => {
            ctl.unloadModelFn(model.metadata.name, true)
            ctl.loadModelFn(model.metadata.name, yamlDump(model), true, true)
        })

        ctl.unloadPipelineFn(pipeline.pipelineName, true)
        ctl.loadPipelineFn(pipeline.pipelineName, pipeline.pipelineCRYaml, true, true)

        const seldonRuneTime = generateSeldonRuntime(replicaModelGw, replicaPipeLineGw, replicaDataFlowEngine)
        ctl.loadSeldonRuntimeFn(seldonRuneTime.object, false, true)

        // wait for pipeline to be ready since scaling data-plane services
        awaitPipelineStatus(pipelineName, "Ready")

        return config
    }, {
        "useKubeControlPlane": true,
    })
}

// runScaleDataFlowEngineOpTimes how many times to scale up dataflow-engine during test, it will be evenly distributed across
// the progress of the test
const runScaleDataFlowEngineOpTimes = 3
// vuIDToRunScaleOps VU ID to run scale services operation from
const vuIDToRunScaleOps = 1;
// scaledOpRunHistory records how many times scaled up has run during test
let scaledOpRunHistory = 0;
// replicaCountModelGw how many replicas for dataflow-engine
let replicaCountDataFlowEngine = replicaDataFlowEngine;


export default function (config) {
    if (exec.vu.idInTest === vuIDToRunScaleOps) {
        if (exec.scenario.progress >= ((scaledOpRunHistory + 1) / (runScaleDataFlowEngineOpTimes + 1))) {
            replicaCountDataFlowEngine++;
            console.log("Scaling UP dataflow-engine to " + replicaCountDataFlowEngine + " replicas at " + Math.round(exec.scenario.progress * 100) + "% progress");

            k8s.init()
            const ctl = connectControlPlaneOps(config)
            const seldonRuneTime = generateSeldonRuntime(replicaModelGw, replicaPipeLineGw, replicaCountDataFlowEngine)
            ctl.loadSeldonRuntimeFn(seldonRuneTime.object, false, true)
            scaledOpRunHistory++;
        }
    }

    const inferPayload = getModelInferencePayload(inputModelType, 1)
    inferHttp(config.inferHttpEndpoint, pipelineName, inferPayload.http, true, true, config.debug, null)
}

export function teardown(config) {
    tearDownK6(config, function (config) {
        const ctl = connectControlPlaneOps(config)

        for (const modelName of [inputModelName1, inputModelName2, outputModelName]) {
            console.log("deleting model ", modelName)
            ctl.unloadModelFn(modelName, true)
        }

        console.log("deleting pipeline ", pipelineName)
        ctl.unloadPipelineFn(pipelineName, false)

        ctl.unloadServerFn(serverName, false, false)
    })
}
