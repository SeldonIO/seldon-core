import { getConfig } from '../components/settings.js'
import { doInfer, setupBase, teardownBase, getVersionSuffix, applyModelReplicaChange } from '../components/utils.js'
import { vu } from 'k6/execution';

export const options = {
    thresholds: {
        'http_req_duration{scenario:default}': [`max>=0`],
        'http_reqs{scenario:default}': [],
        'grpc_req_duration{scenario:default}': [`max>=0`],
        'data_received{scenario:default}': [],
        'data_sent{scenario:default}': [],
    },
    scenarios: {
        ramping_request_rates: {
            startTime: '0s',
            executor: 'ramping-arrival-rate',
            startRate: 5,
            timeUnit: '1s',
            preAllocatedVUs: 50, // how large the initial pool of VUs would be
            maxVUs: 1000, // if the preAllocatedVUs are not enough, we can initialize more
            stages: getConfig().rateStages,
        },
    },
    setupTimeout: '6000s',
    teardownTimeout: '6000s',
};

export function setup() {
    const config = getConfig()

    setupBase(config)
    console.log("rate stages:", getConfig().rateStages)

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

    var rest_enabled = Number(config.inferHttpIterations)
    var grpc_enabled = Number(config.inferGrpcIterations)
    if (rest_enabled && grpc_enabled) {
        // if both protocols are enabled, choose one randomly
        const rand = Math.random()
        if (rand > 0.5) {
            doInfer(modelName, modelNameWithVersion, config, true, idx) // rest
        } else {
            doInfer(modelName, modelNameWithVersion, config, false, idx) // grpc
        }
    } else if (rest_enabled) {
        doInfer(modelName, modelNameWithVersion, config, true, idx)
    } else if (grpc_enabled) {
        doInfer(modelName, modelNameWithVersion, config, false, idx)
    } else {
        throw new Error('Both REST and GRPC protocols are disabled!')
    }

    // for simplicity we only change model replicas in the first VU
    if (vu.idInTest == 1 && config.enableModelReplicaChange) {
        applyModelReplicaChange(config)
    }
}

export function teardown(config) {
    teardownBase(config)
}
