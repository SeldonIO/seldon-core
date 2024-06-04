/**
 * This test aims to verify that Core 2 remains in an error-free state after
 * numerous control-plane operations, with some of those being executed in
 * parallel.
 *
 * Tests Create/Update/Delete operations; expect: consistent end-state,
 * consistent states between operator/scheduler at each intermediary check.
 *
 * State consistency tests happen periodically on VU1, while other VUs wait for
 * the checks to complete.
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
 *
 * When testing with pipelines (i.e running `make deploy-kpipeline-test`) we are also
 * testing pipeline creation/deletion. The pipeline operation follows the same pattern
 * for the model operation, with the following differences:
 * There assumptions to note with the current change:
 * - The test that we currently have is for a single model pipelines
 * - To update a pipeline we induce a change to the pipeline CR that has no real effect
 *   (by setting `batch` to a random value)
 * - We delete pipelines 50% of the time when a delete of a model is required,
 *   to simulate a pipeline that is not available due to issues with models.
 */

import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import { sleep, check } from 'k6';
import { scenario, vu, test } from 'k6/execution';
import { Counter } from 'k6/metrics';
import * as k8s from '../components/k8s.js';
import * as scheduler from '../components/scheduler.js'

import { getConfig } from '../components/settings.js'
import { seldonObjectType, seldonOpExecStatus, seldonOpType } from '../components/seldon.js';
import {
    setupBase,
    periodicExclusiveRun,
    checkModelsStateIsConsistent,
    checkPipelinesStateIsConsistent,
    checkExperimentsStateIsConsistent,
} from '../components/utils.js'
import { generateModel, generatePipelineName } from '../components/model.js';

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

// Control-plane operations timing
const maxRandomDelay = 2
const maxOpDuration = k8s.MAX_RETRIES + 2
const maxIterDuration = maxOpDuration + getConfig().k8sDelaySecPerVU + maxRandomDelay

// Variable only used by VU 1, to avoid running the check twice in the same
// period; Pass object to periodicExclusiveRun(), as the status argument.
var checkStatus = {
    isDue: true
}

// Control-plane client handles
var kubeClient = null
var schedClient = null

// metrics
const createCounter = new Counter("ctl_op_model_create")
const createFailCounter = new Counter("ctl_op_model_create_fail")
const updateCounter = new Counter("ctl_op_model_update")
const updateFailCounter = new Counter("ctl_op_model_update_fail")
const deleteCounter = new Counter("ctl_op_model_delete")
const deleteFailCounter = new Counter("ctl_op_model_delete_fail")

export function setup() {
    var config = getConfig()

    // Check config sanity
    if((config.checkStateEverySec - config.maxCheckTimeSec) < (maxIterDuration + 1)) {
        throw new Error(`Invalid config: CHECK_STATE_EVERY_SECONDS - MAX_CHECK_TIME_SECONDS must be at least ${maxIterDuration + 1} to ensure test progression`)
    }

    setupBase(config)
    let numLoadedModels = config.maxNumModels.reduce((s,a) => s + Number(a), 0)
    console.log("Loaded models (end-of-setup):" + numLoadedModels)
    createCounter.add(numLoadedModels)

    return config
}

function handleCtlOp(config, op, modelTypeIx, existingModels, existingPipelines) {
    var modelName = config.modelNamePrefix[modelTypeIx]
    var pipelineName = generatePipelineName(modelName)
    var altPipelineName = pipelineName // used for delete only
    var modelCRYaml = {}
    if (config.isLoadPipeline) {
        var pipelineCRYaml = {}
    }

    // generate model CR or select one of the existing ones as the
    // target for the control-plane operation, possibly updating its config if
    // needed
    switch (op) {
        case seldonOpType.CREATE:
            VU_maxModelId += 1
            // vu.idInTest starts from 1; leave the range [0, config.maxCreateOpsPerVU - 1] to
            // the preloaded models
            modelName += (vu.idInTest * config.maxCreateOpsPerVU) + VU_maxModelId
            const i = modelTypeIx
            let m = generateModel(config.modelType[i], modelName, 1, config.modelReplicas[i], config.isSchedulerProxy, config.modelMemoryBytes[i], config.inferBatchSize[i])
            modelCRYaml = m.modelCRYaml
            if (config.isLoadPipeline) {
                pipelineName = m.pipelineDefn.pipeline.name
                pipelineCRYaml = m.pipelineCRYaml
            }
            break;
        case seldonOpType.UPDATE:
        case seldonOpType.DELETE:
            var randomModelIx = Math.floor(Math.random() * existingModels.length)
            modelName = existingModels[randomModelIx]

            var randomPipelineIx = Math.floor(Math.random() * existingPipelines.length)
            altPipelineName = existingPipelines[randomPipelineIx]

            if (op === seldonOpType.DELETE) {
                break;
            }

            try {
                let model = kubeClient.get(seldonObjectType.MODEL.description, modelName, config.namespace)

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
                if (config.isLoadPipeline) {
                    let pipeline = kubeClient.get(seldonObjectType.PIPELINE.description, pipelineName, config.namespace)
                    let steps = pipeline.spec.steps
                    steps[0]["batch"] = {"size": Math.round(Math.random() * 100)}  // to induce a change in pipeline
                    let newPipelineCRYaml = {
                        "apiVersion": "mlops.seldon.io/v1alpha1",
                        "kind": "Pipeline",
                        "metadata": {
                            "name": pipelineName,
                            "namespace": getConfig().namespace,
                        },
                        "spec": {
                            "steps": steps,
                            "output": pipeline.spec.output
                        }
                    }
                    pipelineCRYaml = yamlDump(newPipelineCRYaml)
                }
            } catch (err) {
                // just continue test, another VU might have deleted the chosen model
                console.log(`Failed to update model ${modelName}: ${err}`)
                return false
            }
            break;
    }

    console.log(`VU ${vu.idInTest} executes ${op.description} on ${modelName}`)

    // execute control-plane operation
    var opOk = seldonOpExecStatus.FAIL
    switch(op) {
        case seldonOpType.CREATE:
        case seldonOpType.UPDATE:
            opOk = k8s.loadModel(modelName, modelCRYaml, true)
            if (opOk === seldonOpExecStatus.OK && config.isLoadPipeline) {
                opOk = k8s.loadPipeline(pipelineName, pipelineCRYaml, true)
            }
            break;
        case seldonOpType.DELETE:
            opOk = k8s.unloadModel(modelName, true)
            if (opOk === seldonOpExecStatus.OK && config.isLoadPipeline) {
                // We don't want to always delete the pipeline corrsponding to
                // the deleted model, because we also want to test the case
                // where the pipeline remains without some of the component
                // models.
                //
                // However, when we don't delete the pipeline associated with
                // the model, we still want to delete a pipeline, because
                // otherwise the number of pipelines will grow indefinitely.
                // We have picked this altPipeline at random from the existing
                // ones. In practice, this means that the probability to delete
                // the pipeline associated with the model is slightly larger than
                // 0.5
                let unloadAltPipeline = Math.random() < 0.5 ? 0 : 1
                if (unloadAltPipeline) {
                    opOk = k8s.unloadPipeline(altPipelineName, true)
                } else {
                    opOk = k8s.unloadPipeline(pipelineName, true)
                }
            }
            break;
    }

    var targetCounter = null
    switch(opOk) {
        case seldonOpExecStatus.OK:
            switch(op) {
                case seldonOpType.CREATE:
                    targetCounter = createCounter
                    break;
                case seldonOpType.UPDATE:
                    targetCounter = updateCounter
                    break;
                case seldonOpType.DELETE:
                    targetCounter = deleteCounter
                    break;
            }
            break;
        case seldonOpExecStatus.FAIL:
            switch(op) {
                case seldonOpType.CREATE:
                    targetCounter = createFailCounter
                    break;
                case seldonOpType.UPDATE:
                    targetCounter = updateFailCounter
                    break;
                case seldonOpType.DELETE:
                    targetCounter = deleteFailCounter
                    break;
            }
            break;

        // Ignore opOk == seldonOpExecStatus.CONCURRENT_OP_FAIL case;
        // It just means another VU did an operation invalidating the one of the
        // current VU, concurrently; example: current VU wants to update a model,
        // another VU already deleted it. It is expected that such operations
        // would fail, and they shouldn't be recorded as errors as long as the
        // control plane remains functional
    }

    if (targetCounter != null) {
        targetCounter.add(1)
    }

    return opOk
}

export default function (config) {
    kubeClient = k8s.init()
    if(config.enableStateCheck) {
        if (periodicExclusiveRun(config.checkStateEverySec,
                                config.maxCheckTimeSec,
                                maxIterDuration, checkStatus)) {
            console.log(`VU ${vu.idInTest} starts a state consistency check...`)
            // Perform state consistency checks
            let k8sModels = k8s.getAllModels()

            if (schedClient === null) {
                schedClient = scheduler.connectScheduler(config.schedulerEndpoint)
            }

            // The folowing code all executes asynchronously; It's the only way
            // we can currently get grpc status streaming from the scheduler.
            //
            // The VU will return immediately, but all the async functions below
            // are registered to run as part of the event loop, and k6 will wait
            // some time for them to complete before ending the current iteration
            scheduler.getAllModels().then((schedModels) => {
                let msc = checkModelsStateIsConsistent(k8sModels, schedModels)

                if (config.isLoadPipeline) {
                    let k8sPipelines = k8s.getAllPipelines()
                    scheduler.getAllPipelines().then((schedPipelines) => {
                        let psc = checkPipelinesStateIsConsistent(k8sPipelines, schedPipelines)

                        if (!psc && config.stopOnCheckFailure) {
                            test.abort("Aborting test due to pipeline state inconsistencies")
                        }
                    })
                }

                if (config.loadExperiment) {
                    let k8sExperiments = k8s.getAllExperiments()
                    scheduler.getAllExperiments().then((schedExperiments) => {
                        let esc = checkExperimentsStateIsConsistent(k8sExperiments, schedExperiments)

                        if (!esc && config.stopOnCheckFailure) {
                            test.abort("Aborting test due to experiment state inconsistencies")
                        }
                    })
                }

                if (!msc && config.stopOnCheckFailure) {
                    test.abort("Aborting test due to model state inconsistencies")
                }
            })

            // required for the async operations set above to run before the next
            // iteration starts
            return
        }
    }

    const numModelTypes = config.modelType.length

    let candidateIdxs = []
    for (let i = 0; i < numModelTypes; i++) {
        if (config.maxNumModels[i] !== 0)
            candidateIdxs.push(i)
    }
    const numCandidates = candidateIdxs.length

    var idx = candidateIdxs[Math.floor(Math.random() * numCandidates)]
    let modelsOfType = k8s.getExistingModelNames(config.modelNamePrefix[idx])
    let pipelinesOfType = k8s.getExistingPipelineNames(config.modelNamePrefix[idx])

    // Pick random operation in CREATE/UPDATE/DELETE, amongst available ones.
    // Each operation is added multiple times in the selection array as
    // configured in the `createUpdateDeleteBias` array. This defines the
    // probability ratios between the operations. For example, the [1,4,3]
    // createUpdateDeleteBias array makes an Update four times more likely then
    // a Create, and a Delete 3 times more likely than the Create.
    //
    // Because each VU picks the operation independently, it's possible to
    // temporarily get more models than MAX_NUM_MODELS for a given model type.
    let availableOps = []
    if (modelsOfType.length == 0 && config.createUpdateDeleteBias[0] == 0 ) {
        // Force single create operation when maxNumModels for a given type
        // is not 0, all models of that type have been deleted, but the "Create"
        // bias is also set to 0
        availableOps.push(seldonOpType.CREATE)
    }
    if (modelsOfType.length < config.maxNumModels[idx] + config.maxNumModelsHeadroom[idx]) {
        let createArray = Array(config.createUpdateDeleteBias[0]).fill(seldonOpType.CREATE)
        availableOps.push(...createArray)
    }
    if (modelsOfType.length > 0) {
        let updateArray = Array(config.createUpdateDeleteBias[1]).fill(seldonOpType.UPDATE)
        let deleteArray = Array(config.createUpdateDeleteBias[2]).fill(seldonOpType.DELETE)
        availableOps.push(...updateArray)
        availableOps.push(...deleteArray)
    }

    if (availableOps.length == 0) {
        console.log(`No available operations for models of type ${config.modelNamePrefix[idx]}`)
        return
    }
    const randomOp = availableOps[Math.floor(Math.random() * availableOps.length)]
    const opOk = handleCtlOp(config, randomOp, idx, modelsOfType, pipelinesOfType)

    if (opOk === seldonOpExecStatus.OK) {
        sleep(config.k8sDelaySecPerVU + (Math.random() * 2))
    } else {
        // prevent client-rate limiting in case of error
        sleep(2)
    }
}

export function teardown(config) {
    let modelNames = k8s.getExistingModelNames()
    for (var modelName in modelNames) {
        k8s.unloadModel(modelName, false)
    }
    scheduler.disconnectScheduler()
}
