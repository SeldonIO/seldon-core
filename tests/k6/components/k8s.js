import { Kubernetes } from "k6/x/kubernetes";
import { getConfig } from '../components/settings.js'
import {
  awaitStatus,
  awaitPipelineStatus,
  awaitExperimentStart,
  awaitExperimentStop
} from '../components/scheduler.js';
import { seldonObjectType } from '../components/seldon.js'

const kubeclient = new Kubernetes();
const namespace = getConfig().namespace;
var schedulerClient = null;

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
    } catch(error) {
        return false
    }
}

export function loadModel(modelName, data, awaitReady=true) {
    // TODO: Update existing model with new CR definition.
    // At the moment, if an object with the same name exists, it will not be
    // re-loaded with different settings. This is because we get a k8s apply
    // conflict caused by a FieldManager being set on `.spec.memory`
    if(!seldonObjExists(seldonObjectType.MODEL, modelName, namespace)) {
        kubeclient.apply(data)
        let created = kubeclient.get(seldonObjectType.MODEL.description, modelName, namespace)
        if ('uid' in created.metadata) {
            if (awaitReady && schedulerClient != null) {
                awaitStatus(modelName, "ModelAvailable")
            }
        }
    }
}

export function unloadModel(modelName, awaitReady=true) {
    if(seldonObjExists(seldonObjectType.MODEL, modelName, namespace)) {
        kubeclient.delete(seldonObjectType.MODEL.description, modelName, namespace)
        if (awaitReady && schedulerClient != null) {
            awaitStatus(modelName, "ModelTerminated")
        }
    }
}

export function loadPipeline(pipelineName, data, awaitReady=true) {
    if(!seldonObjExists(seldonObjectType.PIPELINE, pipelineName, namespace)) {
        kubeclient.apply(data)
        let created = kubeclient.get(seldonObjectType.PIPELINE.description, pipelineName, namespace)
        if ('uid' in created.metadata) {
            if (awaitReady && schedulerClient != null) {
                awaitStatus(pipelineName, "PipelineReady")
            }
        }
    }
}

export function unloadPipeline(pipelineName, awaitReady = true) {
    if(seldonObjExists(seldonObjectType.PIPELINE, pipelineName, namespace)) {
        kubeclient.delete(seldonObjectType.PIPELINE.description, pipelineName, namespace)
        if (awaitReady && schedulerClient != null) {
            awaitStatus(pipelineName, "PipelineTerminated")
        }
    }
}

export function loadExperiment(experimentName, data, awaitReady=true) {
    if(!seldonObjExists(seldonObjectType.EXPERIMENT, experimentName, namespace)) {
        kubeclient.apply(data)
        let created = kubeclient.get(seldonObjectType.EXPERIMENT.description, experimentName, namespace)
        if ('uid' in created.metadata) {
            if (awaitReady && schedulerClient != null) {
                awaitExperimentStart(experimentName)
            }
        }
    }
}

export function unloadExperiment(experimentName, awaitReady=true) {
    if(seldonObjExists(seldonObjectType.EXPERIMENT, experimentName, namespace)) {
        kubeclient.delete(seldonObjExists.EXPERIMENT.description, experimentName, namespace)
        if (awaitReady && schedulerClient != null) {
            awaitExperimentStop(experimentName)
        }
    }
}
