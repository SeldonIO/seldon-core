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
        return __ENV.MODEL_TYPE
    }
    return "iris"
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
        return __ENV.MAX_NUM_MODELS
    }
    return 10
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
        return __ENV.MODEL_MEMORY_BYTES
    }
    return null
}

function inferBatchSize() {
    if (__ENV.INFER_BATCH_SIZE) {
        return parseInt(__ENV.INFER_BATCH_SIZE)
    }
    return 1
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

function isLoadPipeline() {
    if (__ENV.DATAFLOW_TAG) {
        return !(__ENV.DATAFLOW_TAG === "")
    }
    return false
}

function dataflowTag() {
    if (__ENV.DATAFLOW_TAG) {
        return __ENV.DATAFLOW_TAG
    }
    return ""  // empty means that we should not go via kafka
}

function modelNamePrefix() {
    if (__ENV.MODELNAME_PREFIX) {
        return __ENV.MODELNAME_PREFIX
    }
    return "model"
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

function inferType() {
    if (__ENV.INFER_TYPE) {
        return __ENV.INFER_REST
    }
    return "REST"
}

function doWarmup() {
    if (__ENV.WARMUP) {
        return (__ENV.WARMUP === "true")
    }
    return true
}

export function getConfig() {
    return {
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
        "isLoadPipeline": isLoadPipeline(),
        "dataflowTag": dataflowTag(),
        "modelNamePrefix": modelNamePrefix(),
        "experimentNamePrefix": experimentNamePrefix(),
        "loadExperiment" : loadExperiment(),
        "unloadExperiment": unloadExperiment(),
        "modelStartIdx" : modelStartIdx(),
        "modelEndIdx" : modelEndIdx(),
        "modelName" : modelName(),
        "inferType" : inferType(),
        "doWarmup": doWarmup(),
    }
}