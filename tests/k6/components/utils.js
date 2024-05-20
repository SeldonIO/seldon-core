import { sleep } from 'k6';
import { generateModel, generatePipelineName } from '../components/model.js';
import { connectScheduler,
  disconnectScheduler,
  loadModel,
  unloadModel,
  loadPipeline,
  unloadPipeline,
  awaitPipelineStatus,
  loadExperiment,
  unloadExperiment
} from '../components/scheduler.js';
import {
    connectScheduler as connectSchedulerProxy,
    disconnectScheduler as disconnectSchedulerProxy,
    loadModel as loadModelProxy,
    unloadModel as unloadModelProxy
} from '../components/scheduler_proxy.js';
import { seldonObjectType } from '../components/seldon.js'
import { inferGrpcLoop, inferHttpLoop, modelStatusHttp } from '../components/v2.js';
import * as k8s from '../components/k8s.js';

export function setupBase(config) {
    if (config.loadModel) {
        const ctl = connectControlPlaneOps(config)

        for (let j = 0; j < config.maxNumModels.length; j++) {
            for (let i = 0; i < config.maxNumModels[j]; i++) {
                const modelName = config.modelNamePrefix[j] + i.toString()
                const model = generateModel(config.modelType[j], modelName, 1, config.modelReplicas[j], config.isSchedulerProxy, config.modelMemoryBytes[j], config.inferBatchSize[j])

                var defs = getSeldonObjDef(config, model, seldonObjectType.MODEL)

                ctl.loadModelFn(modelName, defs.model.modelDefn, true)
                if (config.isLoadPipeline) {
                    ctl.loadPipelineFn(generatePipelineName(modelName), defs.model.pipelineDefn, false)  // we use pipeline name as model name
                }
            }
        }

        // note: this doesnt work in case of kafka
        if (!config.isLoadPipeline) {
            for (let j = 0; j < config.maxNumModels.length; j++) {
                const n = config.maxNumModels[j] - 1
                if (n >= 0) {
                  const modelName = config.modelNamePrefix[j] + n.toString()
                  const modelNameWithVersion = modelName + getVersionSuffix(config.isSchedulerProxy)  // first version
                  while (modelStatusHttp(config.inferHttpEndpoint, config.isEnvoy?modelName:modelNameWithVersion, config.isEnvoy) !== 200) {
                      sleep(1)
                  }
                }
            }
        } else {
            for (let j = 0; j < config.maxNumModels.length; j++) {
                const n = config.maxNumModels[j] - 1
                if (n >= 0) {
                  const modelName = config.modelNamePrefix[j] + n.toString()
                  awaitPipelineStatus(generatePipelineName(modelName), "PipelineReady")
                }
            }
        }

        if (config.doWarmup) {
            // warm up
            for (let j = 0; j < config.maxNumModels.length; j++) {
                for (let i = 0; i < config.maxNumModels[j]; i++) {
                    const modelName = config.modelNamePrefix[j] + i.toString()

                    const modelNameWithVersion = modelName + getVersionSuffix(config.isSchedulerProxy)  // first version

                    const model = generateModel(config.modelType[j], modelNameWithVersion, 1, 1, config.isSchedulerProxy, config.modelMemoryBytes[j])

                    inferHttpLoop(
                        config.inferHttpEndpoint, config.isEnvoy?modelName:modelNameWithVersion, model.inference.http, 1, config.isEnvoy, config.dataflowTag)
                }
            }
        }

        disconnectControlPlaneOps(ctl, config)
    }
}

export function teardownBase(config ) {
    if (config.unloadModel) {
        const ctl = connectControlPlaneOps(config)

        for (let j = 0; j < config.maxNumModels.length; j++) {
            for (let i = 0; i < config.maxNumModels[j]; i++) {
                const modelName = config.modelNamePrefix[j] + i.toString()
                // if we have added a pipeline, unloaded it
                if (config.isLoadPipeline) {
                    ctl.unloadPipelineFn(generatePipelineName(modelName))
                }

                ctl.unloadModelFn(modelName, false)
            }
        }

        disconnectControlPlaneOps(ctl, config)
    }
}

function warnFn(fnName, cause, name, data, awaitReady=true) {
  console.log("WARN: "+ fnName + " function not implemented." + cause)
}

export function connectControlPlaneOps(config) {
  var ctl = {}

  ctl.connectSchedulerFn = connectScheduler
  ctl.disconnectSchedulerFn = disconnectScheduler
  if (config.isSchedulerProxy) {
      ctl.connectSchedulerFn = connectSchedulerProxy
      ctl.disconnectSchedulerFn = disconnectSchedulerProxy
  }

  if (config.useKubeControlPlane) {
    ctl.loadModelFn = k8s.loadModel
    ctl.unloadModelFn = k8s.unloadModel
    ctl.loadPipelineFn = k8s.loadPipeline
    ctl.unloadPipelineFn = k8s.unloadPipeline
    ctl.loadExperimentFn = k8s.loadExperiment
    ctl.unloadExperimentFn = k8s.unloadExperiment
  } else {
    ctl.loadModelFn = loadModel
    ctl.unloadModelFn = unloadModel
    ctl.loadPipelineFn = loadPipeline
    ctl.unloadPipelineFn = unloadPipeline
    ctl.loadExperimentFn = loadExperiment
    ctl.unloadExperimentFn = unloadExperiment
    if (config.isSchedulerProxy) {
        const warnCause = "Using SchedulerProxy"
        ctl.loadModelFn = loadModelProxy
        ctl.unloadModelFn = unloadModelProxy
        ctl.loadPipelineFn = warnFn.bind(this, "loadPipeline", warnCause)
        ctl.unloadPipelineFn = warnFn.bind(this, "unloadPipeline", warnCause)
        ctl.loadExperimentFn = warnFn.bind(this, "loadExperiment", warnCause)
        ctl.unloadExperimentFn = warnFn.bind(this, "unloadExperiment", warnCause)
    }
  }

  const schedClient = ctl.connectSchedulerFn(config.schedulerEndpoint)
  // pass scheduler client to k8s for Model/Pipeline status queries
  if (config.useKubeControlPlane && !config.isSchedulerProxy) {
    k8s.init()
    k8s.connectScheduler(schedClient)
  }

  return ctl
}

export function disconnectControlPlaneOps(ctl, config) {
  if (config.useKubeControlPlane && !config.isSchedulerProxy) {
    k8s.disconnectScheduler()
  }
  ctl.disconnectSchedulerFn()
}

export function getSeldonObjDef(config, object, type) {
  var objDef = {
    "model": {
      "modelDefn": null,
      "pipelineDefn": null,
    },
    "experiment": {
      "model1Defn": null,
      "model2Defn": null,
      "experimentDefn": null
    }
  };

  if (config.useKubeControlPlane) {
    switch (type) {
      case seldonObjectType.MODEL:
        objDef.model.modelDefn = object.modelCRYaml
        objDef.model.pipelineDefn = object.pipelineCRYaml
        break;
      case seldonObjectType.EXPERIMENT:
        objDef.experiment.model1Defn = object.model1CRYaml
        objDef.experiment.model2Defn = object.model2CRYaml
        objDef.experiment.experimentDefn = object.experimentCRYaml
        break;
    }
  } else {
    switch (type) {
      case seldonObjectType.MODEL:
        objDef.model.modelDefn = object.modelDefn
        objDef.model.pipelineDefn = object.pipelineDefn
        break;
      case seldonObjectType.EXPERIMENT:
        objDef.experiment.model1Defn = object.model1Defn
        objDef.experiment.model2Defn = object.model2Defn
        objDef.experiment.experimentDefn = object.experimentDefn
        break;
    }
  }

  return objDef
}

export function doInfer(modelName, modelNameWithVersion, config, isHttp, idx) {
    const model = generateModel(config.modelType[idx], modelName, 1, 1, config.isSchedulerProxy, config.modelMemoryBytes[idx], config.inferBatchSize[idx])
    const httpEndpoint = config.inferHttpEndpoint
    const grpcEndpoint = config.inferGrpcEndpoint

    if (config.infer) {
        if (isHttp) {
            inferHttpLoop(
                httpEndpoint, config.isEnvoy?modelName:modelNameWithVersion, model.inference.http, config.inferHttpIterations, config.isEnvoy, config.dataflowTag)
        } else {
            inferGrpcLoop(
                grpcEndpoint, config.isEnvoy?modelName:modelNameWithVersion, model.inference.grpc, config.inferGrpcIterations, config.isEnvoy, config.dataflowTag)
        }
    }
}

export function getVersionSuffix(isSchedulerProxy) {
    var versionSuffix = "_1"
    if (isSchedulerProxy) {
        versionSuffix = "_0"
    }
    return versionSuffix
}
