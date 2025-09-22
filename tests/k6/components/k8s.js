import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import { Kubernetes } from "k6/x/kubernetes";
import { getConfig } from '../components/settings.js'
import {
  awaitExperimentStart,
  awaitExperimentStop
} from '../components/scheduler.js';
import { seldonObjectType, seldonOpExecStatus } from '../components/seldon.js'
import { sleep } from 'k6';

const seldon_target_ns = getConfig().namespace;
export const MAX_RETRIES = 10;
var kubeclient = null;
var schedulerClient = null;

export function init() {
    // We want to initialize/reinitialize this every time init() is called,
    // (rather than only doing it when kubeclient is null) because if k8s
    // operations happen within VUs, on starting a new iteration, the internal
    // kubeclient state becomes invalid (all operations will return
    // client-side throttling errors). This does mean each VU will need to call
    // init() at the beginning of the iteration.
    kubeclient = new Kubernetes();
    return kubeclient
}

export function connectScheduler(schedulerCl) {
  schedulerClient = schedulerCl
}

export function disconnectScheduler() {
  schedulerClient = null
}

function seldonObjExists(kind, name, ns) {
    // This is ugly, but xk6-kubernetes kubeclient.get(...) throws an exception if the
    // underlying k8s CR doesn't exist.

    // The alternative here would be to list all objects of the given kind from the namespace
    // and see if the one with the specified name exists among them. However, that would end
    // up being considerably slower, and we don't want to do it on every single
    // model/pipeline/experiment load or unload.
    try {
        kubeclient.get(kind.description, name, ns)
        return true
    } catch(_) {
        return false
    }
}

function getPrefixAndSuffixFilter(prefix, suffix) {
  let filterFn = null
  if(prefix !== "" || suffix !== ""){
      filterFn = (name) => name.startsWith(prefix) && name.endsWith(suffix)
  }
  return filterFn
}

function getSeldonObjectList(type, mapFn=null, filterFn=null) {
    try {
        let objList = kubeclient.list(type.description, seldon_target_ns)
        if (mapFn !== null) {
            objList = objList.map(mapFn)
        }
        if (filterFn !== null) {
            objList = objList.filter(filterFn)
        }
        return objList
    } catch (error) {
        console.log(`K8S error in listing ${type.description}: ${error}`)
        return []
    }
}

function getObjectCondition(objCR, targetConditionType, field = null) {
    var k8sCondition = {
        "value": "K8sStatusUnknown",
        "met": false
    };
    if('status' in objCR) {
        let status = objCR.status
        if ('conditions' in status) {
            for (let i = 0; i < status.conditions.length; i++){
                let condition = objCR.status.conditions[i]
                if (condition.type === targetConditionType) {
                    if (field === null) {
                        k8sCondition.value = condition
                    } else {
                        k8sCondition.value = condition[field]
                    }
                    k8sCondition.met = (condition.status === "True")
                    break
                }
            }
        }
    }
    return k8sCondition
}


/************
 *
 * Models
 *
 *****/

export function getAllModels() {
    return getSeldonObjectList(seldonObjectType.MODEL)
}

/**
 * getExistingModelNames() can be used to get the models currently loaded in the
 * configured namespace.
 *
 * When passing prefix/suffix constrains, a filtered list is returned.
 *
 * With suitable model names (i.e consistent naming per type, distinct prefixes
 * amongst types), passing a namePrefix is used to filter models of a given type
 *
 * If each VU creating models uses a consistent naming scheme, appending an ID
 * from a range assigned to that specific VU, nameSuffix may be used to retrieve
 * all the models created by the current VU.
 *
 * In combination, it is suggested that namePrefix and nameSuffix may be used
 * together to retrieve all models of a given type that are managed by a given
 * VU
 */
export function getExistingModelNames(namePrefix="", nameSuffix="") {
    try {
        const filterFn = getPrefixAndSuffixFilter(namePrefix, nameSuffix)
        const modelNames = getSeldonObjectList(
            seldonObjectType.MODEL,
            (modelCR) => modelCR.metadata.name,
            filterFn
        )
        return modelNames
    } catch (error) {
        console.log("K8S List Models Error:" + error)
        return []
    }
}

export function modelConditionMet(modelName, targetCondition) {
    var modelObj = kubeclient.get(seldonObjectType.MODEL.description, modelName, seldon_target_ns)
    return getObjectCondition(modelObj, targetCondition, "status").met
}

export function getModelReadyCondition(modelCR) {
    return getObjectCondition(modelCR, "ModelReady", "message")
}

export function loadModel(modelName, data, awaitReady=true, throwError=false) {
    try {
        kubeclient.apply(data)
        let created = kubeclient.get(seldonObjectType.MODEL.description, modelName, seldon_target_ns)
        if ('uid' in created.metadata) {
            if (awaitReady) {
                return awaitStatus(modelName, "ModelReady")
            }
        }
        return seldonOpExecStatus.OK
    } catch (error) {
        if (throwError) {
            throw error
        }
        // continue on error. the apply may be concurrent with a delete and fail
        return seldonOpExecStatus.CONCURRENT_OP_FAIL
    }
}

export function awaitStatus(modelName, status) {
    let retries = 0
    try {
        while (!modelConditionMet(modelName, status)) {
            sleep(1)
            retries++
            if(retries > MAX_RETRIES) {
                console.log(`Giving up on waiting for model ${modelName} to reach status ${status}, after ${MAX_RETRIES} retries`)
                return seldonOpExecStatus.FAIL
            }
        }
        return seldonOpExecStatus.OK
    } catch (_) {
        // in case getModelStatus throws an exception
        return seldonOpExecStatus.CONCURRENT_OP_FAIL
    }
}

export function unloadModel(modelName, awaitReady=true, throwError=false) {
    if(seldonObjExists(seldonObjectType.MODEL, modelName, seldon_target_ns)) {
        try {
            kubeclient.delete(seldonObjectType.MODEL.description, modelName, seldon_target_ns)
            if (awaitReady) {
                let retries = 0
                while (seldonObjExists(seldonObjectType.MODEL, modelName, seldon_target_ns)) {
                    sleep(1)
                    retries++
                    if(retries > MAX_RETRIES) {
                        console.log(`Failed to unload model ${modelName} after ${MAX_RETRIES}, giving up`)
                        return seldonOpExecStatus.FAIL
                    }
                }
            }
            return seldonOpExecStatus.OK
        } catch(error) {
            if (throwError) {
                throw error
            }
            // catch case where model was deleted concurrently by another VU
        }
    }
    return seldonOpExecStatus.CONCURRENT_OP_FAIL
}


/************
 *
 * Pipelines
 *
 *****/

export function getAllPipelines() {
    return getSeldonObjectList(seldonObjectType.PIPELINE)
}

/**
 * getExistingPipelineNames() can be used to get the pipelines currently loaded in the
 * configured namespace.
 *
 * When passing prefix/suffix constrains, a filtered list is returned.
 *
 * With suitable pipeline names (i.e consistent naming per type, distinct prefixes
 * amongst types), passing a namePrefix is used to filter pipelines of a given type
 */
export function getExistingPipelineNames(namePrefix="", nameSuffix="") {
    try {
        const filterFn = getPrefixAndSuffixFilter(namePrefix, nameSuffix)
        const pipelineNames = getSeldonObjectList(
            seldonObjectType.PIPELINE,
            (pipelineCR) => pipelineCR.metadata.name,
            filterFn
        )
        return pipelineNames
    } catch (error) {
        console.log("K8S List Pipelines Error:" + error)
        return []
    }
}

export function pipelineConditionMet(pipelineName, targetCondition) {
    let pipelineObj = kubeclient.get(seldonObjectType.PIPELINE.description, pipelineName, seldon_target_ns)
    return getObjectCondition(pipelineObj, targetCondition, "status").met
}

export function serverConditionMet(serverName, targetCondition) {
    let serverObj = kubeclient.get(seldonObjectType.SERVER.description, serverName, seldon_target_ns)
    return getObjectCondition(serverObj, targetCondition, "status").met
}

export function seldonRuntimeConditionMet(targetCondition) {
    let seldonRuntimeObj = kubeclient.get("seldonruntime.mlops.seldon.io", getConfig().seldonRuntimeName, seldon_target_ns)
    return getObjectCondition(seldonRuntimeObj, targetCondition, "status").met
}

export function getPipelineReadyCondition(pipelineCR) {
    return getObjectCondition(pipelineCR, "PipelineReady", "reason")
}

export function loadPipeline(pipelineName, data, awaitReady=true, throwError=false) {
    try {
        kubeclient.apply(data)
        let created = kubeclient.get(seldonObjectType.PIPELINE.description, pipelineName, seldon_target_ns)
        if ('uid' in created.metadata) {
            if (awaitReady) {
                return awaitPipelineStatus(pipelineName, "PipelineReady")
            }
        }
        return seldonOpExecStatus.OK
    } catch (error) {
        if (throwError) {
            throw error
        }
        // continue on error. the apply may be concurrent with a delete and fail
        return seldonOpExecStatus.CONCURRENT_OP_FAIL
    }
}

export function awaitPipelineStatus(pipelineName, status) {
    let retries = 0
    try {
        while (!pipelineConditionMet(pipelineName, status)) {
            sleep(1)
            retries++
            if(retries > MAX_RETRIES) {
                console.log(`Giving up on waiting for pipeline ${pipelineName} to reach status ${status}, after ${MAX_RETRIES}`)
                return seldonOpExecStatus.FAIL
            }
        }
        return seldonOpExecStatus.OK
    } catch(_) {
        // in case getPipelineStatus throws an exception
        return seldonOpExecStatus.CONCURRENT_OP_FAIL
    }
}

export function awaitServerStatus(serverName, status, throwError=false, maxRetries = 10) {
    let retries = 0
    try {
        while (!serverConditionMet(serverName, status)) {
            sleep(1)
            retries++
            if(retries > maxRetries) {
                const msg = `Giving up on waiting for server ${serverName} to reach status ${status}, after ${maxRetries}`
                if (throwError) {
                    throw msg
                }
                console.log(msg)
                return seldonOpExecStatus.FAIL
            }
        }
        return seldonOpExecStatus.OK
    } catch(error) {
        if (throwError) {
            throw error
        }
        return seldonOpExecStatus.CONCURRENT_OP_FAIL
    }
}

export function awaitSeldonRuntime(retries = 0, throwError = false) {
    try {
        for (const condition of ['DataflowEngineReady', 'EnvoyReady', 'HodometerReady', 'ModelGatewayReady', 'PipelineGatewayReady','SchedulerReady']) {
            if (!seldonRuntimeConditionMet(condition)) {
                sleep(1)
                retries++
                if(retries > MAX_RETRIES) {
                    console.log(`Giving up on waiting for SeldonRuntime ${condition} to reach status True, after ${MAX_RETRIES}`)
                    return seldonOpExecStatus.FAIL
                }
                return awaitSeldonRuntime(retries)
            }
        }
    } catch(error) {
        if (throwError) {
            throw error
        }
        return seldonOpExecStatus.FAIL
    }

    return seldonOpExecStatus.OK
}

export function unloadPipeline(pipelineName, awaitReady = true, throwError=false) {
    if(seldonObjExists(seldonObjectType.PIPELINE, pipelineName, seldon_target_ns)) {
        try {
            kubeclient.delete(seldonObjectType.PIPELINE.description, pipelineName, seldon_target_ns)
            if (awaitReady) {
                let retries = 0
                while (seldonObjExists(seldonObjectType.PIPELINE, pipelineName, seldon_target_ns)) {
                    sleep(1)
                    retries++
                    if(retries > MAX_RETRIES) {
                        console.log(`Failed to unload pipeline ${pipelineName} after ${MAX_RETRIES}, giving up`)
                        return seldonOpExecStatus.FAIL
                    }
                }
            }
            return seldonOpExecStatus.OK
        } catch(error) {
            if (throwError) {
                throw error
            }
            // catch case where model was deleted concurrently by another VU
        }
    }
    return seldonOpExecStatus.CONCURRENT_OP_FAIL
}


/************
 *
 * Experiments
 *
 *****/

export function getAllExperiments() {
    return getSeldonObjectList(seldonObjectType.EXPERIMENT)
}

export function loadExperiment(experimentName, data, awaitReady=true) {
    kubeclient.apply(data)
    let created = kubeclient.get(seldonObjectType.EXPERIMENT.description, experimentName, seldon_target_ns)
    if ('uid' in created.metadata) {
        if (awaitReady && schedulerClient != null) {
            awaitExperimentStart(experimentName)
        }
    }
}

export function unloadExperiment(experimentName, awaitReady=true) {
    if(seldonObjExists(seldonObjectType.EXPERIMENT, experimentName, seldon_target_ns)) {
        kubeclient.delete(seldonObjExists.EXPERIMENT.description, experimentName, seldon_target_ns)
        if (awaitReady && schedulerClient != null) {
            awaitExperimentStop(experimentName)
        }
    }
}


/************
 *
 * SeldonConfig
 *
 *****/


export function generateSeldonConfig(dataFlowEngineMemoryLimit = "1G") {
    let seldonconfig = kubeclient.get("seldonconfig.mlops.seldon.io", getConfig().seldonConfigName, seldon_target_ns)

    seldonconfig.spec.components.forEach(component => {
        if (component.name === "seldon-dataflow-engine") {
            component.podSpec.containers[0].resources = {
                limits: {
                    memory: dataFlowEngineMemoryLimit,
                },
                requests: {
                    cpu: "500m",
                    memory: "1G"
                }
            }
        }
    });

    return {
        "object" : seldonconfig,
        "yaml" : yamlDump(seldonconfig)
    }
}

/************
 *
 * SeldonRuntime
 *
 *****/


export function generateSeldonRuntime(modelGwReplicas, pipelineGwReplicas, dataFlowEngineReplicas, dataFlowEngineMemoryLimit = "1G") {
    let runtime = kubeclient.get("seldonruntime.mlops.seldon.io", getConfig().seldonRuntimeName, seldon_target_ns)

    const updatedReplicas = {
        spec : {
            seldonConfig: "default",
            overrides : [
                {
                    name: "hodometer",
                    replicas: 1,
                },
                {
                    name: "seldon-scheduler",
                    replicas: 1,
                    serviceType: "LoadBalancer"
                },
                {
                    name: "seldon-envoy",
                    replicas: 1,
                    serviceType: "LoadBalancer"
                },
                {
                    name: "seldon-dataflow-engine",
                    replicas: dataFlowEngineReplicas,
                    podSpec: {
                        containers: [{
                            name: "seldon-dataflow-engine",
                            resources: {
                                limits: {
                                    memory: dataFlowEngineMemoryLimit,
                                },
                                requests: {
                                    cpu: "500m",
                                    memory: "1G"
                                }
                            }
                        }],
                    },
                },
                {
                    name: "seldon-modelgateway",
                    replicas: modelGwReplicas,
                },
                {
                    name: "seldon-pipelinegateway",
                    replicas: pipelineGwReplicas,
                }
            ]
        }
    }

    const seldonRuntimeSpec = { ...runtime, ...updatedReplicas };

    return {
        "object" : seldonRuntimeSpec,
        "yaml" : yamlDump(seldonRuntimeSpec)
    }
}

export function loadSeldonConfig(data, throwError = false) {
    try {
        kubeclient.update(data)
        return seldonOpExecStatus.OK
    } catch (error) {
        if (throwError) {
            throw error
        }
        return seldonOpExecStatus.FAIL
    }
}

export function loadSeldonRuntime(data, awaitReady=true, throwError=false) {
    try {
        kubeclient.update(data)
        if (awaitReady) {
            return awaitSeldonRuntime(10, true)
        }
        return seldonOpExecStatus.OK
    } catch (error) {
        if (throwError) {
            throw error
        }
        return seldonOpExecStatus.FAIL
    }
}

/************
 *
 * Server
 *
 *****/


export function generateServer(serverName, serverConfig, replicas, minReplicas, maxReplicas) {
    const serverResource = {
        apiVersion: "mlops.seldon.io/v1alpha1",
        kind: "Server",
        metadata: {
            name: serverName,
            namespace: seldon_target_ns,
        },
        spec: {
            maxReplicas: maxReplicas,
            minReplicas: minReplicas,
            replicas: replicas,
            serverConfig: serverConfig,
            statefulSetPersistentVolumeClaimRetentionPolicy: {
                whenDeleted: "Retain",
                whenScaled: "Retain"
            }
        }
    };

    return {
        "object" : serverResource,
        "yaml" : yamlDump(serverResource)
    }
}


export function loadServer(data, serverName, awaitReady=true, throwError=false, maxRetires = 10) {
    try {
        kubeclient.apply(data)
        if (awaitReady) {
            return awaitServerStatus(serverName, "Ready", throwError, maxRetires)
        }
        return seldonOpExecStatus.OK
    } catch (error) {
        if (throwError) {
            throw error
        }
        return seldonOpExecStatus.FAIL
    }
}

export function unloadServer(serverName, awaitReady=true, throwError=false, maxRetries = MAX_RETRIES) {
    if(seldonObjExists(seldonObjectType.SERVER, serverName, seldon_target_ns)) {
        try {
            // check pods have actually deleted
            //
            // let serverPods = kubeclient.get("Pod", )

            kubeclient.delete(seldonObjectType.SERVER.description, serverName, seldon_target_ns)
            if (awaitReady) {
                let retries = 0
                while (seldonObjExists(seldonObjectType.SERVER, serverName, seldon_target_ns)) {
                    sleep(1)
                    retries++
                    if(retries > maxRetries) {
                        const msg = `Failed to unload server ${serverName} after ${MAX_RETRIES}, giving up`
                        if (throwError) {
                            throw msg
                        }
                        console.log(msg)
                        return seldonOpExecStatus.FAIL
                    }
                }
            }
            return seldonOpExecStatus.OK
        } catch(error) {
            if (throwError) {
                throw error
            }
            // catch case where model was deleted concurrently by another VU
        }
    }
}