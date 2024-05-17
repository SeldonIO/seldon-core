import { Kubernetes } from "k6/x/kubernetes";
import { getConfig } from '../components/settings.js'
import {
  awaitExperimentStart,
  awaitExperimentStop
} from '../components/scheduler.js';
import { seldonObjectType } from '../components/seldon.js'
import { sleep } from 'k6';

const seldon_target_ns = getConfig().namespace;
const MAX_RETRIES = 60;
var kubeclient = null;
var schedulerClient = null;

export function init() {
    // We want to initialize/reinitialize this every time init() is called,
    // (rather than only doing it when kubeclient is null) because if k8s
    // operations happen within VUs, on starting a new iteration, the internal
    // kubeclient state becomes invalid (all operations will return
    // client-side throttling errors). This does meant each VU will need to call
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

/**
 * getModels() can be used to get the models currently loaded in the configured
 * namespace.
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
export function getModels(namePrefix="", nameSuffix="") {
    try {
        const modelList = kubeclient.list(seldonObjectType.MODEL.description, seldon_target_ns)
        const modelNames = modelList.map((modelCR) => modelCR.metadata.name)
        if(namePrefix === "" && nameSuffix === ""){
            return modelNames
        }
        const filteredModels = modelNames.filter((s) => s.startsWith(namePrefix) && s.endsWith(nameSuffix))
        return filteredModels
    } catch (error) {
        console.log("K8S List Models Error:" + error)
        return []
    }
}

export function loadModel(modelName, data, awaitReady=true) {
    try {
        kubeclient.apply(data)
        let created = kubeclient.get(seldonObjectType.MODEL.description, modelName, seldon_target_ns)
        if ('uid' in created.metadata) {
            if (awaitReady) {
                return awaitStatus(modelName, "ModelReady")
            }
        }
        return true
    } catch (_) {
        // continue on error. the apply may be concurrent with a delete and fail
        return false
    }
}

export function awaitStatus(modelName, status) {
    let retries = 0
    try {
        while (!getModelStatus(modelName, status)) {
            sleep(1)
            retries++
            if(retries > MAX_RETRIES) {
                console.log(`Giving up on waiting for model ${modelName} to reach status ${status}, after ${MAX_RETRIES}`)
                return false
            }
        }
        return true
    } catch (_) {
        // in case getModelStatus throws an exception
        return false
    }
}

export function getModelStatus(modelName, targetStatus) {
    var modelObj = kubeclient.get(seldonObjectType.MODEL.description, modelName, seldon_target_ns)
    if ('status' in modelObj) {
        let status = modelObj.status
        if ('conditions' in status) {
            for (let i = 0; i < status.conditions.length; i++) {
                if(status.conditions[i].type === targetStatus) {
                    return status.conditions[i].status === "True"
                }
            }
        }
    }

    return false
}

export function unloadModel(modelName, awaitReady=true) {
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
                        return false
                    }
                }
            }
            return true
        } catch(_) {
            // catch case where model was deleted concurrently by another VU
        }
    }
    return false
}

export function loadPipeline(pipelineName, data, awaitReady=true) {
    try {
        kubeclient.apply(data)
        let created = kubeclient.get(seldonObjectType.PIPELINE.description, pipelineName, seldon_target_ns)
        if ('uid' in created.metadata) {
            if (awaitReady) {
                return awaitPipelineStatus(pipelineName, "PipelineReady")
            }
        }
        return true
    } catch (_) {
        // continue on error. the apply may be concurrent with a delete and fail
        return false
    }
}

export function awaitPipelineStatus(pipelineName, status) {
    let retries = 0
    try {
        while (!getPipelineStatus(pipelineName, status)) {
            sleep(1)
            retries++
            if(retries > MAX_RETRIES) {
                console.log(`Giving up on waiting for pipeline ${pipelineName} to reach status ${status}, after ${MAX_RETRIES}`)
                return false
            }
        }
        return true
    } catch(_) {
        // in case getPipelineStatus throws an exception
        return false
    }
}

export function getPipelineStatus(pipelineName, targetStatus) {
    let pipelineObj = kubeclient.get(seldonObjectType.PIPELINE.description, pipelineName, seldon_target_ns)
    if ('status' in pipelineObj) {
        let status = pipelineObj.status
        if ('conditions' in status) {
            for (let i = 0; i < status.conditions.length; i++) {
                // return the most recent state (true/false) of the condition
                // type named <targetStatus>
                if(status.conditions[i].type == targetStatus) {
                    return status.conditions[i].status === "True"
                }
            }
        }
    }
    return false
}

export function unloadPipeline(pipelineName, awaitReady = true) {
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
                        return false
                    }
                }
            }
            return true
        } catch(_) {
            // catch case where model was deleted concurrently by another VU
        }
    }
    return false
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
