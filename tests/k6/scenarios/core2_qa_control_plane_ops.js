/**
 * This test aims to verify that Core 2 remains in an error-free state after
 * numerous control-plane operations, with some of those being executed in
 * parallel.
 *
 * Tests Create/Update/Delete operations; expect: consistent end-state
 *
 * Constant VU test;
 * Per VU:
 * 1. Pick arbitrary model type
 * 2. Pick arbitrary **available** operation from the `Create/Update/Delete` set
 *      - `Create` not available if max number of models of that type already loaded
 *      - `Delete` not available if no models of that type remain
 *      - If `Update` is picked, choose one of the existing models of that type
 *        and apply some random variation to the model’s memory requirements
 *        (+/- MAX_UPDATE_MEM ratio) and/or number of replicas
 *        (+/- up to the number of config.maxModelReplicas)
 * 3. Apply operation to cluster
 * 4. Wait VU_OP_DELAY_SECONDS
 * 5. Repeat from 1
 */

import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import { sleep } from 'k6';
import { vu } from 'k6/execution';
import { Counter } from 'k6/metrics';
import * as k8s from '../components/k8s.js';

import { getConfig } from '../components/settings.js'
import { seldonObjectType, seldonOpType } from '../components/seldon.js';
import { setupBase } from '../components/utils.js'
import { generateModel } from '../components/model.js';

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
    duration: '30m',
    teardownTimeout: '6000s',
}

// Each VU gets a range of Ids for the models it creates, not overlapping
// with other VUs. The range size is controlled by MAX_CREATE_OPS_PER_VU
var VU_maxModelId = 0

const createCounter = new Counter("ctl_op_model_create")
const updateCounter = new Counter("ctl_op_model_update")
const deleteCounter = new Counter("ctl_op_model_delete")

var kubeclient = {}

export function setup() {
    var config = getConfig()

    setupBase(config)
    console.log(config.maxNumModels)
    let numLoadedModels = config.maxNumModels.reduce((s,a) => s + Number(a), 0)
    console.log("Loaded models (end-of-setup):" + numLoadedModels)
    createCounter.add(numLoadedModels)

    return config
}

function handleCtlOp(config, op, modelTypeIx, existingModels) {
    var modelName = config.modelNamePrefix[modelTypeIx]
    var modelCRYaml = {}

    // generate model CR or select one of the existing ones as the
    // target for the control-plane operation, possibly updating its config if
    // needed
    switch (op) {
        case seldonOpType.CREATE:
            VU_maxModelId += 1
            modelName += (vu.idInTest * config.maxCreateOpsPerVU) + VU_maxModelId
            const i = modelTypeIx
            let m = generateModel(config.modelType[i], modelName, 1, config.modelReplicas[i], config.isSchedulerProxy, config.modelMemoryBytes[i], config.inferBatchSize[i])
            modelCRYaml = m.modelCRYaml
            break;
        case seldonOpType.UPDATE:
        case seldonOpType.DELETE:
            var randomModelIx = Math.floor(Math.random() * existingModels.length)
            modelName = existingModels[randomModelIx]

            if (op === seldonOpType.DELETE) {
                break;
            }

            try {
                let model = kubeclient.get(seldonObjectType.MODEL.description, modelName, config.namespace)

                let plusOrMinus = Math.random() < 0.5 ? -1 : 1
                // update memory +/- with random variation
                let mem = parseInt(config.modelMemoryBytes[modelTypeIx], 10)
                let memUnit = model.spec.memory.substring(String(mem).length)
                let memVariation = Math.round(Math.random() * mem * config.maxMemUpdateFraction)
                let newMemory = String(mem + (memVariation * plusOrMinus)) + memUnit
                // update replicas +/- with random variation
                let replicasToMax = config.maxModelReplicas[modelTypeIx] - model.spec.replicas
                var deltaReplicas = plusOrMinus > 0 ? replicasToMax : model.spec.replicas
                let replicasVariation = Math.round(Math.random() * deltaReplicas)
                let newReplicas = model.spec.replicas + (replicasVariation * plusOrMinus)
                if (newReplicas < 1) newReplicas = 1

                let newModelCR = {
                    "apiVersion": "mlops.seldon.io/v1alpha1",
                    "kind": "Model",
                    "metadata": {
                        "name": modelName,
                        "namespace": config.namespace
                    },
                    "spec": {
                        "storageUri": model.spec.storageUri,
                        "requirements": model.spec.requirements,
                        "memory": newMemory,
                        "replicas": newReplicas
                    }
                }
                modelCRYaml = yamlDump(newModelCR)
            } catch (_) {
                // just continue test, another VU might have deleted the chosen model
                return false
            }
            break;
    }

    console.log(`VU ${vu.idInTest} executes ${op.description} on ${modelName}`)


    // execute control-plane operation
    var opOk = false
    switch(op) {
        case seldonOpType.CREATE:
        case seldonOpType.UPDATE:
            opOk = k8s.loadModel(modelName, modelCRYaml, true)
            break;
        case seldonOpType.DELETE:
            opOk = k8s.unloadModel(modelName, true)
            break;
    }

    // update k6 metrics
    if (opOk === true) {
        switch(op) {
            case seldonOpType.CREATE:
                createCounter.add(1)
                break;
            case seldonOpType.UPDATE:
                updateCounter.add(1)
                break;
            case seldonOpType.DELETE:
                deleteCounter.add(1)
                break;
        }
    }

    return opOk

}

export default function (config) {
    kubeclient = k8s.init()
    const numModelTypes = config.modelType.length

    var idx = Math.floor(Math.random() * numModelTypes)
    while (config.maxNumModels[idx] === 0) {
        idx = Math.floor(Math.random() * numModelTypes)
    }
    let modelsOfType = k8s.getModels(config.modelNamePrefix[idx])

    // Pick random operation in CREATE/UPDATE/DELETE, amongst available ones.
    // Each operation is added multiple times in the selection array as
    // configured in the `createUpdateDeleteBias` array. This defines the
    // probability ratios between the operations. For example, the [1,4,3]
    // createUpdateDeleteBias array makes an Update four times more likely then
    // a Create, and a Delete 3 times more likely than the Create.
    //
    // Because each VU picks the operation independently, it's possible to
    // temporarily get more models than MAX_NUM_MODELS for a given model type.
    let availableOps = Array(config.createUpdateDeleteBias[1]).fill(seldonOpType.UPDATE)
    if (modelsOfType.length < config.maxNumModels[idx] + config.maxNumModelsHeadroom[idx]) {
        let createArray = Array(config.createUpdateDeleteBias[0]).fill(seldonOpType.CREATE)
        availableOps.push(...createArray)
    }
    if (modelsOfType.length > 0) {
        let deleteArray = Array(config.createUpdateDeleteBias[2]).fill(seldonOpType.DELETE)
        availableOps.push(...deleteArray)
    }
    const randomOp = availableOps[Math.floor(Math.random() * availableOps.length)]
    handleCtlOp(config, randomOp, idx, modelsOfType)

    sleep(config.k8sDelaySecPerVU + (Math.random() * 2))

}

export function teardown(config) {
    let modelNames = k8s.getModels()
    for (var modelName in modelNames) {
        k8s.unloadModel(modelName, false)
    }
}
