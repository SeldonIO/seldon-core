import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import * as k8s from '../../../components/k8s.js';
import * as scheduler from '../../../components/scheduler.js'
import { sleep } from 'k6';

import { getConfig } from '../../../components/settings.js'
import {
    connectControlPlaneOps, doInfer, getVersionSuffix,
} from '../../../components/utils.js'
import {generateMultiModelPipelineYaml, getModelInferencePayload} from '../../../components/model.js';
import {generateSeldonRuntime} from "../../../components/k8s.js";

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
const modelName = 'tests-pipeline-1-node-echo-1';

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

    const replicaModelGw = 2;
    const replicaDataFlowEngine = 1;
    const replicaPipeLineGw = 1;


    const seldonRuneTime = generateSeldonRuntime(replicaModelGw,replicaPipeLineGw,replicaDataFlowEngine)

    ctl.loadSeldonRuntimeFn(seldonRuneTime.object, true, true)
    // TODO appears to be bug in seldon operator where it doesn't set the status=false for the various components while
    //  they're scaling up/down... we sleep until fixed
    sleep(5)

    return config
}

export default function (config) {
    const modelNameWithVersion = "_2"
    doInfer(modelName, modelName + modelNameWithVersion, config, false, 0)
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
