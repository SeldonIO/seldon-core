import { sleep } from 'k6';
import { scenario, vu } from 'k6/execution';
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
import { seldonObjectType, getSeldonObjectCommonName } from '../components/seldon.js'
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
                    ctl.loadPipelineFn(generatePipelineName(modelName), defs.model.pipelineDefn, true)  // we use pipeline name as model name + "-pipeline"
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

/**
 * periodicExclusiveRun(...) allows setting up k6 iteration code such that,
 * periodically and for a fixed period, a single VU (VU 1) executes while other
 * VUs wait. This is not a great synchronization primitive because the
 * implementation relies on timestamps and sleeping. However, we choose this
 * strategy as there is no implicit communication between VUs and we don't want
 * to introduce external mechanisms or setup requirements (additional http
 * endpoints to call in order to sync, a database, etc).
 *
 * However, for avoiding concurrency issues under the chosen design we need
 * fairly large margins (say, 0.5 to 1s) for the following parameters:
 *  - maxIterDuration, the max duration of an iteration (including any waits)
 *  - maxRuntimeSec, the max duration of the function running exclusively
 * In particular, the guarantees of exclusive execution might not hold towards
 * the end of the intervals defined by those durations.
 *
 * @param {Number} exclusiveEverySec - The period where VU 1 runs exclusively
 * starts every exclusiveEverySec seconds.
 * @param {Number} maxRuntimeSec - The maximum time that the exclusive run might
 * take. It's important for the exclusive code to complete well within this time;
 * other VUs will only wait for this long before continuing, so after
 * maxRuntimeSec VU1 looses execution exclusivity even if it hasn't finished
 * running its "exclusive" code.
 * @param {Number} maxIterDuration - The maximum time that a normal
 * (non-exclusive) iteration might take. This is used to determine if there is
 * enough time left before the next exclusive interval for another iteration to
 * run. If any iterations run longer than this duration, we can no longer
 * guarantee that the VU1 code run when periodicExclusiveRun(...) returns true
 * will run exclusively.
 * @param {object} status - The status object for the exclusive run. During the
 * first call, this should be status = { isDue: true }. VU 1 needs to pass
 * the same object here across iterations (a global) to keep track of whether
 * the periodic exclusive run is due and to avoid running it twice in the same
 * period.
 * @returns {boolean} - Returns true only for VU 1, at the point where it is in
 * the right time slot for running the exclusive check.
 *
 * Every VU should call periodicExclusiveRun(...) at the beginning of each
 * iteration. The same configuration arguments need to be passed on each call
 * (there's no support to modify the exclusive run frequency or max runtime).
 *
 * - If there is sufficient time for the VU to execute another iteration,
 * periodicExclusiveRun simply returns false, and the iteration of the VU
 * calling periodicExclusiveRun should continue as normal.
 * - If not, VU1 waits until the time when the next exclusive run is scheduled
 * and returns true. Other VUs wait until the time of the next exclusive run +
 * the maximum time that the exclusive run might take, then return false.
 *
 * Example usage:
 *
 *   // globals
 *   const status = { isDue: true }
 *   const exclusiveEverySec = 30
 *   const maxRuntimeSec = 5
 *   const maxIterDuration = 10
 *   ....
 *   export default function (data) { // iteration code
 *      if (periodicExclusiveRun(exclusiveEverySec, maxRuntimeSec, maxIterDuration, status)) {
 *         // Code exclusively run on VU 1 every exclusiveEverySec seconds
 *         // while other VUs wait
 *
 *         // return recommended if the exclusive run code sets any async
 *         // operations; this will allow those to execute on the event loop
 *         // before the next iteration starts
 *         return
 *      }
 *
 *      // Normal iteration code (parallel, on all VUs)
 *   }
 *
 * The following diagram (not to scale) shows periods when iterations may run
 * (marked with I) vs times when all VUs except the first one wait for the check
 * on VU 1 to complete. It also shows, for an arbitrary iteration,
 * the timeToNextRun as used in the implementation to determine if enough time
 * exists to run another iteration:
 *
 * |<------------------exclusiveEverySec------------------>|<----exclusiveEverySec-----..
 * |<---maxRuntimeSec--->|           |<--maxIterDuration-->|<---maxRuntimeSec--->|     ..
 * |  VU 1 runs check    | I I I I I I                     |  VU 1 runs check    | I I ..
 * |  other VUs wait     |       ^-----timeToNextRun------>|  other VUs wait     |
 */
export function periodicExclusiveRun(exclusiveEverySec, maxRuntimeSec, maxIterDuration, status) {
  let timeSoFar = new Date().getTime() - scenario.startTime
  const exclusiveEveryMs = exclusiveEverySec * 1000
  const timeToNextRun = (exclusiveEveryMs - (timeSoFar % exclusiveEveryMs)) / 1000.0

  // The case when a VU ends up calling periodicExclusiveRun during the time
  // reserved for the exclusive run. It will occur on test startup (first
  // exclusive run), and defend against iterations starting too soon in other
  // cases.
  const deltaFromPeriodStart = exclusiveEverySec - timeToNextRun
  if (deltaFromPeriodStart < maxRuntimeSec) {
      if (vu.idInTest == 1) {
          if(status.isDue === true) {
              status.isDue = false
              return true
          }
      } else {
          sleep(maxRuntimeSec - deltaFromPeriodStart)
      }
      return false
  }

  // We've passed beyond maxExclusiveRuntimeSec within the current period.
  // Reset checkStatus for VU1, preparing for the next period
  if (vu.idInTest == 1) {
      status.isDue = true
  }
  // Do we have time to run another iteration (operation)?
  if (timeToNextRun < maxIterDuration) {
      // no time for another iteration; VU 1 will wait for the exclusive run
      // time slot, others will wait until the exclusive run interval ends.
      if (vu.idInTest == 1) {
          console.log("VU 1 now waiting for the exclusive run time slot")
          sleep(timeToNextRun)
          status.isDue = false
          return true
      } else {
          console.log(`VU ${vu.idInTest} now waiting for the exclusive run interval to end`)
          sleep(timeToNextRun + maxRuntimeSec)
      }
  }
  return false
}

/**
 * Basic state consistency checks that apply to all seldon objects (Models,
 * Pipelines, Experiments).
 *
 * There are 3 high-level checks implemented by this function:
 *
 * 1. The number of objects of the given type is the same in K8S and the scheduler
 * 2. All objects in K8S are in the scheduler
 * 3. There are no additional models present as non-terminated in the scheduler
 * compared to the ones reported by K8S
 *
 * Custom checks when comparing two objects with the same name can be added by
 * providing a perMatchCheckFn function. Other custom checks should be implemented
 * separately.
 *
 * @param {*} k8sObjects - the array of objects (CRs) from Kubernetes.
 * @param {*} schedObjects - the array of status objects as returned by the scheduler
 * @param {*} objType - one of the seldon.seldonObjectType values
 * @param {*} perMatchCheckFn - user-provided function running for each matched
 * object pair (K8S, Scheduler). It should return true if the object states are
 * consistent, and false if they are not. The function should accept three
 * arguments: the K8S object, the Scheduler object, and an array to which it can
 * append any inconsistencies found: function perMatchCheckFn(k8sObject,
 * schedObject, inconsistencies)
 * @returns {object} - Returns an object with the following properties:
 * - schedObjIndex: the scheduler objects of the given type, indexed by name
 * - stateIsConsistent: true if the state is consistent, false otherwise
 * - inconsistencies: an array of inconsistencies found
 *
 */
function checkSeldonObjectStateIsConsistent(k8sObjects, schedObjects, objType, perMatchCheckFn = null) {
  let statesConsistent = true
  let inconsistencies = []
  let objName = getSeldonObjectCommonName(objType)

  // Filter-out scheduler state objects that have been terminated, as those
  // will never appear on the k8s side
  let activeSchedObjs = schedObjects.filter(obj => {
    switch (objType) {
      case seldonObjectType.MODEL:
        return obj.versions[0].state.state !== "ModelTerminated"
      case seldonObjectType.PIPELINE:
        return obj.versions[0].state.status !== "PipelineTerminated"
      case seldonObjectType.EXPERIMENT:
        return !obj.active
    }
  });

  // Build index for objects currently active in the scheduler
  let schedObjsIndex = {}
  for (let i = 0; i < activeSchedObjs.length; i++) {
      let propName = objName.one.charAt(0).toLowerCase() + objName.one.slice(1) + "Name"
      let objInstanceName = activeSchedObjs[i][propName]
      schedObjsIndex[objInstanceName] = activeSchedObjs[i].versions[0]
      schedObjsIndex[objInstanceName].found = false
  }

  // CHECK 1: Number of objects in K8S and Scheduler are the same
  if (k8sObjects.length !== activeSchedObjs.length) {
      statesConsistent = false
      let inconsistency = {
          type: `NumberOf${objName.many}`,
          message: `K8S has ${k8sObjects.length} ${objName.many}, Scheduler has ${activeSchedObjs.length} ${objName.many}`
      }
      inconsistencies.push(inconsistency)
  } else {
      console.log(`State consistency check: number of ${objName.many} = ${k8sObjects.length}`)
  }

  // CHECK 2: All objects in K8S are in the Scheduler, delegate per-object
  // state checks to the perMatchCheckFn function, if provided
  for (let i = 0; i < k8sObjects.length; i++) {
      let k8sObject = k8sObjects[i]
      let k8sObjectName = k8sObject.metadata.name
      if (k8sObjectName in schedObjsIndex) {
          let schedObject = schedObjsIndex[k8sObjectName]
          schedObjsIndex[k8sObjectName].found = true

          // More checks applicable for any two seldon objects with the same
          // name can be added here

          if (perMatchCheckFn != null) {
            statesConsistent &= perMatchCheckFn(k8sObject, schedObject, inconsistencies)
          }
      } else {
          statesConsistent = false
          let inconsistency = {
              type: `${objName.one}NotFound`,
              message: `${objName.one} ${k8sObjectName} not found in scheduler`
          }
          inconsistencies.push(inconsistency)
      }
  }

  // CHECK 3: The only models present as non-terminated in the scheduler are
  // the ones we've just checked based on the K8S list
  Object.keys(schedObjsIndex).forEach((objIxName) => {
    if (!schedObjsIndex[objIxName].found) {
        statesConsistent = false
        let inconsistency = {
            type: `${objName.one}NotFound`,
            message: `${objName.one} ${objIxName} not found in K8S but present in scheduler`
        }
        inconsistencies.push(inconsistency)
    }
  })

  return {
    "schedObjIndex": schedObjsIndex,
    "stateIsConsistent": statesConsistent,
    "inconsistencies": inconsistencies,
  }
}

/**
 * Checks if the models' state is consistent between Kubernetes (K8S) and the scheduler.
 *
 * @param {Array} k8sModels - The array of model (CRs) from Kubernetes. This is typically
 * fetched first, via k8sModels = k8s.getAllModels().
 * @param {Array} schedModels - The array of models from the scheduler. This is typically
 * fetched via scheduler.getAllModels().then((schedModels) => {
 *    // call checkModelsStateIsConsistent(k8sModels, schedModels) here
 * }).
 * @returns {boolean} - Returns true if the models' state is consistent, false otherwise.
 */
export function checkModelsStateIsConsistent(k8sModels, schedModels) {
  let statesConsistent = true
  let inconsistencies = []

  // Common checks for all seldon objects plus a number of model-specific checks
  let commonChecks = checkSeldonObjectStateIsConsistent(k8sModels,
    schedModels, seldonObjectType.MODEL,
    (k8sModel, schedModel, inconsArr) => {
      let k8sModelName = k8sModel.metadata.name
      let isConsistent = true

      // check same status & state
      const {value: k8sModelState, met: modelIsReady} = k8s.getModelReadyCondition(k8sModel)
      if (k8sModelState !== schedModel.state.state) {
          isConsistent = false
          let inconsistency = {
              type: "ModelState",
              message: `Model ${k8sModelName} has state ${k8sModelState} in K8S and ${schedModel.state.state} in Scheduler`
          }
          inconsArr.push(inconsistency)
      }

      // if status is "Ready" check same number of replicas
      if (modelIsReady && k8sModel.status.replicas !== schedModel.state.availableReplicas) {
        isConsistent = false
        let inconsistency = {
            type: `ModelReplicas`,
            message: `Model ${k8sModelName} has ${k8sModel.status.replicas} replicas in K8S and ${schedModel.state.availableReplicas} in Scheduler`
        }
        inconsArr.push(inconsistency)
      }

      // if status is "Ready" check if on the same version
      let schedVersion = schedModel.version
      if (modelIsReady && k8sModel.metadata.generation !== schedVersion) {
          isConsistent = false
          let inconsistency = {
          type: `ModelVersion`,
          message: `Model ${k8sModelName} has version ${k8sModel.metadata.generation} in K8S and ${schedVersion} in Scheduler`
          }
          inconsistencies.push(inconsistency)
      }

      return isConsistent
    })

  statesConsistent = commonChecks.stateIsConsistent
  inconsistencies.push(...commonChecks.inconsistencies)

  // Any further checks go here
  // ...

  if (!statesConsistent) {
      console.log("[MODELS] State consistency check: FAIL; inconsistencies:")
      for (let i = 0; i < inconsistencies.length; i++) {
          console.log(inconsistencies[i].message)
      }
  } else {
      console.log("[MODELS] State consistency check: OK")
  }

  return statesConsistent
}

/**
 * Checks if the pipelines' state is consistent between Kubernetes (K8S) and the scheduler.
 *
 * @param {Array} k8sPipelines - The array of pipeline (CRs) from Kubernetes. This is typically
 * fetched first, via k8sPipelines = k8s.getAllPipelines().
 * @param {Array} schedPipelines - The array of pipelines from the scheduler. This is typically
 * fetched via scheduler.getAllPipelines().then((schedPipelines) => {
 *   // call checkPipelinesStateIsConsistent(k8sPipelines, schedPipelines) here
 * }).
 * @returns {boolean} - Returns true if the pipelines' state is consistent, false otherwise.
 */
export function checkPipelinesStateIsConsistent(k8sPipelines, schedPipelines) {
  let statesConsistent = true
  let inconsistencies = []

  // Common checks for all seldon objects
  let commonChecks = checkSeldonObjectStateIsConsistent(
    k8sPipelines, schedPipelines, seldonObjectType.PIPELINE,
    (k8sPipeline, schedPipeline, inconsArr) => {
      let k8sPipelineName = k8sPipeline.metadata.name
      let isConsistent = true

      // check same status
      const {value: k8sPipelineState, met: pipelineIsReady} = k8s.getPipelineReadyCondition(k8sPipeline)
      if (k8sPipelineState !== schedPipeline.state.status) {
          isConsistent = false
          let inconsistency = {
              type: "PipelineState",
              message: `Pipeline ${k8sPipelineName} has status ${k8sPipelineState} in K8S and ${schedPipeline.state.status} in Scheduler`
          }
          inconsArr.push(inconsistency)
      }

      // check same version
      let schedVersion = schedPipeline.pipeline.version
      if (pipelineIsReady && k8sPipeline.metadata.generation !== schedVersion) {
          isConsistent = false
          let inconsistency = {
              type: `PipelineVersion`,
              message: `Pipeline ${k8sPipelineName} has version ${k8sPipeline.metadata.generation} in K8S and ${schedVersion} in Scheduler`
          }
          inconsistencies.push(inconsistency)
      }

      return isConsistent
    })

  statesConsistent = commonChecks.stateIsConsistent
  inconsistencies.push(...commonChecks.inconsistencies)

  // Any further checks go here
  // ...

  if (!statesConsistent) {
    console.log("[PIPELINES] State consistency check: FAIL; inconsistencies:")
    for (let i = 0; i < inconsistencies.length; i++) {
        console.log(inconsistencies[i].message)
    }
  } else {
      console.log("[PIPELINES] State consistency check: OK")
  }

  return statesConsistent
}

export function checkExperimentsStateIsConsistent(k8sExperiments, schedExperiments) {
  return true // to implement
}