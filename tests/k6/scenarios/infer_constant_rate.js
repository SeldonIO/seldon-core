import { getConfig } from '../components/settings.js'
import { doInfer, setupBase, teardownBase, getVersionSuffix, applyModelReplicaChange } from '../components/utils.js'
import { vu } from 'k6/execution';

var kubeClient = null

export const options = {
    thresholds: {
        'http_req_duration{scenario:default}': [`max>=0`],
        'http_reqs{scenario:default}': [],
        'grpc_req_duration{scenario:default}': [`max>=0`],
        'data_received{scenario:default}': [],
        'data_sent{scenario:default}': [],
    },
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: getConfig().requestRate,
            timeUnit: '1s',
            duration: getConfig().constantRateDurationSeconds.toString()+'s',
            preAllocatedVUs: 10, // how large the initial pool of VUs would be
            maxVUs: 1000, // if the preAllocatedVUs are not enough, we can initialize more
        },
    },
    setupTimeout: '6000s',
    teardownTimeout: '6000s',
};

export function setup() {
    const config = getConfig()

    setupBase(config)

    return config
}

export default function (config) {
    const numModelTypes = config.modelType.length

    let candidateIdxs = []
    for (let i = 0; i < numModelTypes; i++) {
        if (config.maxNumModels[i] !== 0)
            candidateIdxs.push(i)
    }
    const numCandidates = candidateIdxs.length

    var idx = candidateIdxs[Math.floor(Math.random() * numCandidates)]
    const modelId = Math.floor(Math.random() * config.maxNumModels[idx])
    const modelName = config.modelNamePrefix[idx] + modelId.toString()

    const modelNameWithVersion = modelName + getVersionSuffix(config.isSchedulerProxy)  // first version

    if (config.inferType === "REST") {
        doInfer(modelName, modelNameWithVersion, config, true, idx)
    } else {
        doInfer(modelName, modelNameWithVersion, config, false, idx)
    }

    // for simplicity we only change model replicas in the first VU
    if (vu.idInTest == 1 && config.enableModelReplicaChange) {
        applyModelReplicaChange(config)
    }
}

export function teardown(config) {
    teardownBase(config)
}
