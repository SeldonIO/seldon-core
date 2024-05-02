import { Kubernetes } from "k6/x/kubernetes";
import { getConfig } from '../components/settings.js'
import {
  awaitStatus,
  awaitPipelineStatus,
  awaitExperimentStart,
  awaitExperimentStop
} from '../components/scheduler.js';

const kubeclient = new Kubernetes();
const namespace = getConfig.namespace;
var schedulerClient = null;

export function connectScheduler(schedulerCl) {
  schedulerClient = schedulerCl
}

export function disconnectScheduler() {
  schedulerClient = null
}

export function loadModel(modelName, data, awaitReady=true) {
    //console.log(data)
    kubeclient.apply(data)
    // let created = kubeclient.get("Model.mlops.seldon.io", modelName, namespace)
    // if ('uid' in created.metadata) {
    if (awaitReady && schedulerClient != null) {
        awaitStatus(modelName, "ModelAvailable")
    }
    // }
}

export function unloadModel(modelName, awaitReady=true) {
    // console.log("Unloading model "+modelName)
    kubeclient.delete("Model.mlops.seldon.io", modelName, namespace)
    if (awaitReady && schedulerClient != null) {
        awaitStatus(modelName, "ModelTerminated")
    }
}

export function loadPipeline(pipelineName, data, awaitReady=true) {
    //console.log(data)
    kubeclient.apply(data)
    // let created = kubeclient.get("Pipeline.mlops.seldon.io", pipelineName, namespace)
    // if ('uid' in created.metadata) {
    if (awaitReady && schedulerClient != null) {
        awaitStatus(pipelineName, "PipelineReady")
    }
    // }
}

export function unloadPipeline(pipelineName, awaitReady = true) {
    kubeclient.delete("Pipeline.mlops.seldon.io", pipelineName, namespace)
    if (awaitReady && schedulerClient != null) {
        awaitStatus(pipelineName, "PipelineTerminated")
    }
}

export function loadExperiment(experimentName, data, awaitReady=true) {
    const ns = data.metadata.namespace
    kubeclient.apply(data)
    let created = kubeclient.get("Experiment.mlops.seldon.io", experimentName, namespace)
    if ('uid' in created.metadata) {
        if (awaitReady && schedulerClient != null) {
            awaitExperimentStart(experimentName)
        }
    }
}

export function unloadExperiment(experimentName, awaitReady=true) {
    kubeclient.delete("Experiment.mlops.seldon.io", experimentName, namespace)
    if (awaitReady && schedulerClient != null) {
        awaitExperimentStop(experimentName)
    }
}
