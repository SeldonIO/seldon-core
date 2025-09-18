function useKubeControlPlane() {
    if (__ENV.USE_KUBE_CONTROL_PLANE) {
        return (__ENV.USE_KUBE_CONTROL_PLANE === "true")
    }
    return false
}

function schedulerEndpoint() {
    if (__ENV.SCHEDULER_ENDPOINT) {
        return __ENV.SCHEDULER_ENDPOINT
    }
    return "0.0.0.0:9004"
}

function inferGrpcEndpoint() {
    if (__ENV.INFER_GRPC_ENDPOINT) {
        return __ENV.INFER_GRPC_ENDPOINT
    }
    return "0.0.0.0:9000"
}

function inferHttpEndpoint() {
    if (__ENV.INFER_HTTP_ENDPOINT) {
        return __ENV.INFER_HTTP_ENDPOINT
    }
    return "http://0.0.0.0:9000"
}

function inferHttpIterations() {
    if (__ENV.INFER_HTTP_ITERATIONS) {
        return __ENV.INFER_HTTP_ITERATIONS
    }
    return 1
}

function inferGrpcIterations() {
    if (__ENV.INFER_GRPC_ITERATIONS) {
        return __ENV.INFER_GRPC_ITERATIONS
    }
    return 1
}

function modelType() {
    if (__ENV.MODEL_TYPE) {
        return __ENV.MODEL_TYPE.split(",")
    }
    return ["iris"]
}

function skipTeardown() {
    if (__ENV.SKIP_TEARDOWN) {
        return true
    }
    return false
}

function skipSetup() {
    if (__ENV.SKIP_SETUP) {
        return true
    }
    return false
}

function loadModel() {
    if (__ENV.SKIP_LOAD_MODEL) {
        return false
    }
    return true
}

function loadExperiment() {
    if (__ENV.SKIP_LOAD_EXPERIMENT) {
        return false
    }
    return true
}

function infer() {
    if (__ENV.SKIP_INFER) {
        return false
    }
    return true
}

function unloadModel() {
    if (__ENV.SKIP_UNLOAD_MODEL) {
        return false
    }
    return true
}

function unloadExperiment() {
    if (__ENV.SKIP_UNLOAD_EXPERIMENT) {
        return false
    }
    return true
}

function maxNumModels() {
    if (__ENV.MAX_NUM_MODELS) {
        return __ENV.MAX_NUM_MODELS.split(",").map(n => Number(n))
    } else if (__ENV.MODEL_TYPE) {
        const num =  __ENV.MODEL_TYPE.split(",").length
        return Array(num).fill(1)
    }
    return [1]
}

function maxNumModelsHeadroom() {
    if (__ENV.MAX_NUM_MODELS_HEADROOM) {
        return __ENV.MAX_NUM_MODELS_HEADROOM.split(",").map(n => Number(n))
    } else if (__ENV.MODEL_TYPE) {
        const num =  __ENV.MODEL_TYPE.split(",").length
        return Array(num).fill(0)
    }
    return [0]
}

function isSchedulerProxy() {
    if (__ENV.SCHEDULER_PROXY) {
        return (__ENV.SCHEDULER_PROXY === "true")
    }
    return false
}

function isEnvoy() {
    if (__ENV.ENVOY) {
        return (__ENV.ENVOY === "true")
    }
    return true
}

function modelMemoryBytes() {
    if (__ENV.MODEL_MEMORY_BYTES) {
        return __ENV.MODEL_MEMORY_BYTES.split(",")
    } else if (__ENV.MODEL_TYPE) {
        const num =  __ENV.MODEL_TYPE.split(",").length
        return Array(num).fill(0)
    }
    return [0]
}

function maxMemUpdateFraction() {
    if (__ENV.MAX_MEM_UPDATE_FRACTION) {
        return Number(__ENV.MAX_MEM_UPDATE_FRACTION)
    }
    return 0
}

function inferBatchSize() {
    if (__ENV.INFER_BATCH_SIZE) {
        return __ENV.INFER_BATCH_SIZE.split(",").map( s => parseInt(s))
    } else if (__ENV.MODEL_TYPE) {
        const num =  __ENV.MODEL_TYPE.split(",").length
        return Array(num).fill(1)
    }
    return [1]
}

function modelReplicas() {
    if (__ENV.MODEL_NUM_REPLICAS) {
        return __ENV.MODEL_NUM_REPLICAS.split(",").map( s => parseInt(s))
    } else if (__ENV.MODEL_TYPE) {
        const num =  __ENV.MODEL_TYPE.split(",").length
        return Array(num).fill(1)
    }
    return [1]
}

function maxModelReplicas() {
    if (__ENV.MAX_MODEL_REPLICAS) {
        return __ENV.MAX_MODEL_REPLICAS.split(",").map( s => parseInt(s))
    } else if (__ENV.MODEL_TYPE) {
        const num =  __ENV.MODEL_TYPE.split(",").length
        return Array(num).fill(1)
    }
    return [1]
}

function createUpdateDeleteBias() {
    if (__ENV.MODEL_CREATE_UPDATE_DELETE_BIAS) {
        return __ENV.MODEL_CREATE_UPDATE_DELETE_BIAS.split(",").map( s => parseInt(s)).slice(0,3)
    }
    return [1, 1, 1]
}

function modelStartIdx() {
    if (__ENV.MODEL_START_IDX) {
        return parseInt(__ENV.MODEL_START_IDX)
    }
    return 0
}

function modelEndIdx() {
    if (__ENV.MODEL_END_IDX) {
        return parseInt(__ENV.MODEL_END_IDX)
    }
    return 0
}

function debug() {
    if (__ENV.K6_DEBUG) {
        return (__ENV.K6_DEBUG === "true")
    }
    return false
}

function viaPipeline() {
    return (__ENV.VIA_PIPELINE === "true")
}

function modelNamePrefix() {
    if (__ENV.MODELNAME_PREFIX) {
        return __ENV.MODELNAME_PREFIX.split(",")
    } else if (__ENV.MODEL_TYPE) {
        return  __ENV.MODEL_TYPE.split(",")
    }
    return ["model"]
}

function modelName() {
    if (__ENV.MODEL_NAME) {
        return __ENV.MODEL_NAME
    }
    return ""
}

function experimentNamePrefix() {
    if (__ENV.EXPERIMENTNAME_PREFIX) {
        return __ENV.EXPERIMENTNAME_PREFIX
    }
    return "experiment"
}

function useGRPC() {
    if (__ENV.USE_GRPC) {
        return true
    }
    return false
}

function seldonRuntimeName() {
    if (__ENV.SELDON_RUNTIME_NAME) {
        return __ENV.SELDON_RUNTIME_NAME
    }
    return "seldon"
}

function requestIDPrefix() {
    if (__ENV.REQUEST_ID_PREFIX) {
        return __ENV.REQUEST_ID_PREFIX
    }
    return null
}

function inferType() {
    if (__ENV.INFER_TYPE) {
        return __ENV.INFER_TYPE
    }
    return "REST"
}

function doWarmup() {
    if (__ENV.WARMUP) {
        return (__ENV.WARMUP === "true")
    }
    return true
}


function requestRate() {
    if (__ENV.REQUEST_RATE) {
        return parseInt(__ENV.REQUEST_RATE)
    }
    return 10
}

function requestRates() {
    if (__ENV.REQUEST_RATES) {
        return __ENV.REQUEST_RATES.split(",").map( s => parseInt(s))
    }
    return [requestRate()]
}

function rateStages() {
    if (__ENV.REQUEST_RATES) {
        var stages = []
        var durations = constantRateDurationsSeconds()
        var rates = requestRates()
        for (var i = 0; i < rates.length; i++) {
            // ramp up (1/3 rd of the duration)
            stages.push({target: rates[i], duration: Math.ceil(durations[i]/3).toString()+'s'})
            // hold
            stages.push({target: rates[i], duration: durations[i].toString()+'s'})
        }
        return stages
    }
    return [{target: requestRate(), duration: constantRateDurationSeconds().toString()+'s'}]
}

function constantRateDurationSeconds() {
    if (__ENV.CONSTANT_RATE_DURATION_SECONDS) {
        return __ENV.CONSTANT_RATE_DURATION_SECONDS
    }
    return 30
}

function constantRateDurationsSeconds() {
    if (__ENV.CONSTANT_RATE_DURATIONS_SECONDS) {
        var durations = __ENV.CONSTANT_RATE_DURATIONS_SECONDS.split(",").map( s => parseInt(s))
        if (durations.length > requestRates().length) {
            return durations.slice(0, requestRates().length)
        } else if (durations.length < requestRates().length) {
            // pad with the last value
            const last = durations[durations.length - 1]
            for (var i = durations.length; i < requestRates().length; i++) {
                durations.push(last)
            }
        } else {
            return durations
        }
    }
    const reqNumberOfStages = requestRates().length
    return new Array(reqNumberOfStages).fill(constantRateDurationSeconds()/reqNumberOfStages)
}

function podNamespace() {
    if (__ENV.NAMESPACE) {
        return __ENV.NAMESPACE
    }
    return "seldon-mesh"
}

function maxCreateOpsPerVU() {
    if (__ENV.MAX_CREATE_OPS_PER_VU) {
        return Number(__ENV.MAX_CREATE_OPS_PER_VU)
    }
    return 10000
}

function k8sDelaySecPerVU() {
    if (__ENV.K8S_DELAY_SECONDS_PER_VU) {
        return Number(__ENV.K8S_DELAY_SECONDS_PER_VU)
    }
    return 10
}

// How often to do state consistency checks (in seconds)
function checkStateEverySec() {
    if (__ENV.CHECK_STATE_EVERY_SECONDS) {
        return Number(__ENV.CHECK_STATE_EVERY_SECONDS)
    }
    return 4 * 60
}

// Maximum time to wait for a state consistency check to complete (in seconds).
// This MUST be fulfilled under all circumstances, otherwise concurrency issues
// will appear. This is because other VUs will start control-plane operations
// after maxCheckTimeSec in the current checkPeriod, irrespective of whether
// the state check is done or not.
function maxCheckTimeSec() {
    if (__ENV.MAX_CHECK_TIME_SECONDS) {
        return Number(__ENV.MAX_CHECK_TIME_SECONDS)
    }
    return 10
}

// Whether to abort the k6 test if a state consistency check fails
function stopOnCheckFailure() {
    if (__ENV.STOP_ON_CHECK_FAILURE) {
        return (__ENV.STOP_ON_CHECK_FAILURE === "true")
    }
    return true
}

function enableStateCheck() {
    if (__ENV.ENABLE_STATE_CHECK) {
        return (__ENV.ENABLE_STATE_CHECK === "true")
    }
    return true
}

function enableModeReplicaChange() {
    if (__ENV.ENABLE_MODEL_REPLICA_CHANGE) {
        return (__ENV.ENABLE_MODEL_REPLICA_CHANGE === "true")
    }
    return false
}

function sleepBetweenModelReplicaChange() {
    if (__ENV.SLEEP_BETWEEN_REPLICA_CHANGE) {
        return Number(__ENV.SLEEP_BETWEEN_REPLICA_CHANGE)
    }
    return 10
}

export function getConfig() {
    return {
        "useGRPC" : useGRPC(),
        "seldonRuntimeName": seldonRuntimeName(),
        "requestIDPrefix" : requestIDPrefix(),
        "skipSetup" : skipSetup(),
        "skipTearDown" : skipTeardown(),
        "debug": debug(),
        "useKubeControlPlane": useKubeControlPlane(),
        "schedulerEndpoint": schedulerEndpoint(),
        "inferHttpEndpoint": inferHttpEndpoint(),
        "inferGrpcEndpoint": inferGrpcEndpoint(),
        "inferHttpIterations": inferHttpIterations(),
        "inferGrpcIterations": inferGrpcIterations(),
        "modelType": modelType(),
        "loadModel": loadModel(),
        "infer": infer(),
        "unloadModel": unloadModel(),
        "maxNumModels": maxNumModels(),
        "isSchedulerProxy": isSchedulerProxy(),
        "isEnvoy": isEnvoy(),
        "modelMemoryBytes": modelMemoryBytes(),
        "inferBatchSize": inferBatchSize(),
        "viaPipeline": viaPipeline(),
        "modelNamePrefix": modelNamePrefix(),
        "experimentNamePrefix": experimentNamePrefix(),
        "loadExperiment" : loadExperiment(),
        "unloadExperiment": unloadExperiment(),
        "modelStartIdx" : modelStartIdx(),
        "modelEndIdx" : modelEndIdx(),
        "modelName" : modelName(),
        "inferType" : inferType(),
        "doWarmup": doWarmup(),
        "requestRate": requestRate(),
        "requestRates": requestRates(),
        "constantRateDurationSeconds": constantRateDurationSeconds(),
        "constantRateDurationsSeconds": constantRateDurationsSeconds(),
        "rateStages": rateStages(),
        "modelReplicas": modelReplicas(),
        "maxModelReplicas": maxModelReplicas(),
        "namespace":  podNamespace(),
        "maxNumModelsHeadroom": maxNumModelsHeadroom(),
        "createUpdateDeleteBias": createUpdateDeleteBias(),
        "maxCreateOpsPerVU": maxCreateOpsPerVU(),
        "k8sDelaySecPerVU": k8sDelaySecPerVU(),
        "maxMemUpdateFraction": maxMemUpdateFraction(),
        "enableStateCheck": enableStateCheck(),
        "checkStateEverySec": checkStateEverySec(),
        "maxCheckTimeSec": maxCheckTimeSec(),
        "stopOnCheckFailure": stopOnCheckFailure(),
        "enableModelReplicaChange": enableModeReplicaChange(),
        "sleepBetweenModelReplicaChange": sleepBetweenModelReplicaChange(),
    }
}
