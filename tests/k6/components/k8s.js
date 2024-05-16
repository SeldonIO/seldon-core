import { Kubernetes } from "k6/x/kubernetes";
import { getConfig } from '../components/settings.js'
import {
  awaitExperimentStart,
  awaitExperimentStop
} from '../components/scheduler.js';
import { seldonObjectType } from '../components/seldon.js'
import { sleep } from 'k6';

const seldon_target_ns = getConfig().namespace;
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

export function getModelsWithNamePrefix(modelNamePrefix) {
    try {
        const modelList = kubeclient.list(seldonObjectType.MODEL.description, seldon_target_ns)
        const modelNames = modelList.map((modelCR) => modelCR.metadata.name)
        if(modelNamePrefix === ""){
            return modelNames
        }
        const filteredModels = modelNames.filter((s) => s.startsWith(modelNamePrefix))
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
            if (awaitReady && schedulerClient != null) {
                awaitStatus(modelName, "ModelReady")
            }
        }
        return true
    } catch (_) {
        // continue on error. the apply may be concurrent with a delete and fail
        return false
    }
}

export function getModelStatus(modelName, targetStatus) {
    try {
        var modelObj = kubeclient.get(seldonObjectType.MODEL.description, modelName, seldon_target_ns)
    } catch (_) {
        // catch the case where the model no longer exists and there is no point in waiting for it
        return true
    }
    if ('status' in modelObj) {
        let status = modelObj.status
        if ('conditions' in status) {
            for (let i = 0; i < status.conditions.length; i++) {
                if(status.conditions[i].type == targetStatus) {
                    return true
                }
            }
            return false
        } else {
            return false
        }
    } else {
        return false
    }
}

export function awaitStatus(modelName, status) {
    while (!getModelStatus(modelName, status)) {
        sleep(1)
    }
}

export function unloadModel(modelName, awaitReady=true) {
    if(seldonObjExists(seldonObjectType.MODEL, modelName, seldon_target_ns)) {
        try {
            kubeclient.delete(seldonObjectType.MODEL.description, modelName, seldon_target_ns)
            if (awaitReady && schedulerClient != null) {
                while (seldonObjExists(seldonObjectType.MODEL, modelName, seldon_target_ns)) {
                    sleep(1)
                }
            }
            return true
        } catch(_) {
            // catch case where model was deleted concurrently by another VU
            return false
        }
    } else {
        return false
    }
}

export function loadPipeline(pipelineName, data, awaitReady=true) {
    kubeclient.apply(data)
    let created = kubeclient.get(seldonObjectType.PIPELINE.description, pipelineName, seldon_target_ns)
    if ('uid' in created.metadata) {
        if (awaitReady && schedulerClient != null) {
            awaitPipelineStatus(pipelineName, "PipelineReady")
        }
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
                    return status.conditions[i].status
                }
            }
            return ""
        } else {
            return ""
        }
    } else {
        return ""
    }
}

export function awaitPipelineStatus(pipelineName, status) {
    while (!getPipelineStatus(pipelineName, status)) {
        sleep(1)
    }
}

export function unloadPipeline(pipelineName, awaitReady = true) {
    if(seldonObjExists(seldonObjectType.PIPELINE, pipelineName, seldon_target_ns)) {
        kubeclient.delete(seldonObjectType.PIPELINE.description, pipelineName, seldon_target_ns)
        if (awaitReady && schedulerClient != null) {
            while (seldonObjExists(seldonObjectType.PIPELINE, pipelineName, seldon_target_ns)) {
                sleep(1)
            }
        }
    }
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
