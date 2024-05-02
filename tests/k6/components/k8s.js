import { Kubernetes } from "k6/x/kubernetes";
import { getConfig } from '../components/settings.js'
import {
  awaitStatus
  awaitPipelineStatus
  awaitExperimentStart
  awaitExperimentStop
} from '../components/scheduler.js';

const kubeclient = new Kubernetes();
var schedulerClient = null;

export function connectScheduler(schedulerCl) {
  schedulerClient = schedulerCl
}

export function disconnectScheduler() {
  schedulerClient = null
}

export function loadModel(modelName, data, awaitReady=true) {
    //console.log(data)
    const ns = data.metadata.namespace
    kubeclient.apply(data)
    let created = kubeclient.get("Model.mlops.seldon.io", modelName, ns)
    if ('uid' in created.metadata) {
        if (awaitReady and schedulerClient != null) {
            awaitStatus(modelName, "ModelAvailable")
        }
    }
}

export function unloadModel(modelName, awaitReady=true) {
    kubeclient.delete("Model.mlops.seldon.io", modelName, getConfig().namespace)
    if (awaitReady and schedulerClient != null) {
        awaitStatus(modelName, "ModelTerminated")
    }
}

export function loadPipeline(pipelineName, data, awaitReady=true) {
    const ns = data.metadata.namespace
    kubeclient.apply(data)
    let created = kubeclient.get("Pipeline.mlops.seldon.io", pipelineName, ns)
    if ('uid' in created.metadata) {
        if (awaitReady and schedulerClient != null) {
            awaitStatus(modelName, "PipelineReady")
        }
    }
}

export function unloadPipeline(pipelineName, awaitReady = true) {
    kubeclient.delete("Pipeline.mlops.seldon.io", pipelineName, getConfig().namespace)
    if (awaitReady and schedulerClient != null) {
        awaitStatus(modelName, "PipelineTerminated")
    }
}

export function loadExperiment(experimentName, data, awaitReady=true) {
    const ns = data.metadata.namespace
    kubeclient.apply(data)
    let created = kubeclient.get("Experiment.mlops.seldon.io", experimentName, ns)
    if ('uid' in created.metadata) {
        if (awaitReady and schedulerClient != null) {
            awaitExperimentStart(experimentName)
        }
    }
}

export function unloadExperiment(experimentName, awaitReady=true) {
    kubeclient.delete("Experiment.mlops.seldon.io", experimentName, getConfig().namespace)
    if (awaitReady and schedulerClient != null) {
        awaitExperimentStop(experimentName)
    }
}
