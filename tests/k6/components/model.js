const tfsimple_string = "tfsimple_string"
const tfsimple = "tfsimple"
const iris = "iris"  // mlserver
const pytorch_cifar10 = "pytorch_cifar10"
const tfmnist = "tfmnist"
const tfresnet152 = "tfresnet152"
const onnx_gpt2 = "onnx_gpt2"
const mlflow_wine = "mlflow_wine" // mlserver
const add10 = "add10" // https://github.com/SeldonIO/triton-python-examples/tree/master/add10
const sentiment = "sentiment" // mlserver

const models = {
    mlflow_wine: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/mlflow/elasticnet_wine_model_mlserver_1_1",
            "maxUriSuffix": 0,
            "requirements": ["mlflow"],
            "memoryBytes": 20000,
        },
    },
    iris: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/testing/iris",
            "maxUriSuffix": 10,
            "requirements": ["sklearn"],
            "memoryBytes": 20000,
        },
    },
    sentiment: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/mlserver/huggingface/sentiment",
            "maxUriSuffix": 0,
            "requirements": ["huggingface"],
            "memoryBytes": 20000,
        },
    },
    tfsimple: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/simple",
            "maxUriSuffix": 0,
            "requirements": ["tensorflow"],
            "memoryBytes": 20000,
        },
    },
    add10: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/add10",
            "maxUriSuffix": 0,
            "requirements": ["tensorflow", "python"],
            "memoryBytes": 20000,
        },
    },
    tfsimple_string: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/tf_simple_string",
            "maxUriSuffix": 0,
            "requirements": ["tensorflow"],
            "memoryBytes": 20000,
        },
    },
    pytorch_cifar10: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/pytorch_cifar10/cifar10",
            "maxUriSuffix": 0,
            "requirements": ["pytorch"],
            "memoryBytes": 20000,
        },
    },
    tfmnist: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/tf_mnist",
            "maxUriSuffix": 0,
            "requirements": ["tensorflow"],
            "memoryBytes": 20000,
        },
    },
    tfresnet152: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/tf_resnet152",
            "maxUriSuffix": 0,
            "requirements": ["tensorflow"],
            "memoryBytes": 20000,
        },
    },
    onnx_gpt2: {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/onnx_gpt2/gpt2",
            "maxUriSuffix": 0,
            "requirements": ["onnx"],
            "memoryBytes": 20000,
        },
    },
}

export function getModelInferencePayload(modelName, inferBatchSize) {
    if (modelName == tfsimple_string) {
        const shape = [inferBatchSize, 16]
        var httpBytes = []
        var grpcBytes = []
        
        for (var i = 0; i < 16 * inferBatchSize; i++) {
            grpcBytes.push("MQ=="); // base64 of 1
            httpBytes.push("97")
        }
        const payload = {
            "http": {"inputs":[{"name":"INPUT0","data":httpBytes,"datatype":"BYTES","shape":shape},{"name":"INPUT1","data":httpBytes,"datatype":"BYTES","shape":shape}]},
            "grpc": {"inputs":[{"name":"INPUT0","contents":{"bytes_contents":grpcBytes},"datatype":"BYTES","shape":shape},{"name":"INPUT1","contents":{"bytes_contents":grpcBytes},"datatype":"BYTES","shape":shape}]}
        }
        return payload
    } else if  (modelName == tfsimple) {
        const shape = [inferBatchSize ,16]
        var data = []
        for (var i = 0; i < 16 * inferBatchSize; i++) {
            data.push(i)
        }
        return {
            "http": {"inputs":[{"name":"INPUT0","data": data,"datatype":"INT32","shape":shape},{"name":"INPUT1","data":data,"datatype":"INT32","shape":shape}]},
            "grpc": {"inputs":[{"name":"INPUT0","contents":{"int_contents":data},"datatype":"INT32","shape":shape},{"name":"INPUT1","contents":{"int_contents":data},"datatype":"INT32","shape":shape}]}
        }
    } else if  (modelName == add10) {
        const shape = [4]
        var data = new Array(4).fill(0.1)
        return {
            "http": {"inputs":[{"name":"INPUT","data": data,"datatype":"FP32","shape":shape}]},
            "grpc": {"inputs":[{"name":"INPUT","contents":{"int_contents":data},"datatype":"FP32","shape":shape}]}
        }
    } else if (modelName == iris) {
        const shape = [inferBatchSize, 4]
        var data = []
        for (var i = 0; i < 4 * inferBatchSize; i++) {
            data.push(i)
        }
        return {
            "http": {"inputs": [{"name": "predict", "shape": shape, "datatype": "FP32", "data": [data]}]},
            "grpc": {"inputs":[{"name":"input","contents":{"fp32_contents":data},"datatype":"FP32","shape":shape}]}
        }
    } else if (modelName == sentiment) {
        const shape = [inferBatchSize]
        var httpBytes = []
        var grpcBytes = []
        const base64 = "dGhpcyBpcyBhIGNvb2wgdGVzdA=="
        const str = "this is a cool test"
        for (var i = 0; i < inferBatchSize; i++) {
            httpBytes.push(str)
            grpcBytes.push(base64)
        }
        return {
            "http": {"inputs": [{"name": "args", "shape": shape, "datatype": "BYTES", "data": httpBytes}]},
            "grpc": {"inputs":[{"name":"args","contents":{"bytes_contents":grpcBytes},"datatype":"BYTES","shape":shape}]}
        }
    } else if (modelName == pytorch_cifar10) {
        const shape = [inferBatchSize, 3, 32, 32]
        const data = new Array(3*32*32*inferBatchSize).fill(0.1)
        const datatype = "FP32"
        return {
            "http": {"inputs":[{"name":"input__0","data": data,"datatype":datatype,"shape":shape}]},
            "grpc": {"inputs":[{"name":"input__0","contents":{"fp32_contents":data},"datatype":datatype,"shape":shape}]}
        }
    } else if (modelName == tfmnist) {
        const shape = [inferBatchSize, 28, 28, 1]
        const data = new Array(28*28*inferBatchSize).fill(0.1)
        const datatype = "FP32"
        return {
            "http": {"inputs":[{"name":"conv2d_input","data": data,"datatype":datatype,"shape":shape}]},
            "grpc": {"inputs":[{"name":"conv2d_input","contents":{"fp32_contents":data},"datatype":datatype,"shape":shape}]}
        }
    } else if (modelName == tfresnet152) {
        const shape = [inferBatchSize, 224, 224, 3]
        const data = new Array(3*224*224*inferBatchSize).fill(0.1)
        const datatype = "FP32"
        return {
            "http": {"inputs":[{"name":"input_1","data": data,"datatype":datatype,"shape":shape}]},
            "grpc": {"inputs":[{"name":"input_1","contents":{"fp32_contents":data},"datatype":datatype,"shape":shape}]}
        }
    } else if (modelName == onnx_gpt2) {
        const shape = [inferBatchSize, 10]
        const data = new Array(10*inferBatchSize).fill(1)
        const datatype = "INT32"
        return {
            "http": {"inputs":[{"name":"input_ids","data": data,"datatype":datatype,"shape":shape}, {"name":"attention_mask","data": data,"datatype":datatype,"shape":shape}]},
            "grpc": {"inputs":[{"name":"input_ids","contents":{"int_contents":data},"datatype":datatype,"shape":shape}, {"name":"attention_mask","contents":{"int_contents":data},"datatype":datatype,"shape":shape}]}
        }
    } else if (modelName == mlflow_wine) {
        const fields = ["fixed acidity", "volatile acidity", "citric acidity", "residual sugar", "chlorides", "free sulfur dioxide", "total sulfur dioxide", "density", "pH", "sulphates", "alcohol"]
        const shape = [1]
        const data = new Array(1).fill(1)
        const data_all = new Array(fields.length).fill(1)
        const datatype = "FP32"
         var v2Fields = [];
        var v2FieldsGrpc = [];
        for (var i = 0; i < fields.length; i++) {
            v2Fields.push({
                "name": fields[i],
                "data": data,
                "datatype": datatype,
                "shape": shape,
            })
            v2FieldsGrpc.push({
                "name": fields[i],
                "contents": {"fp32_contents": data},
                "datatype": datatype,
                "shape": shape,
            })
        }
        return {
            "http": {"inputs": v2Fields, "parameters": {"content_type": "pd"}},
            "grpc": {"inputs": v2FieldsGrpc, "parameters": {"content_type": {"string_param": "pd"}}}
        }
    }
}

export function generateExperiment(experimentName, modelType, modelName1, modelName2, uriOffset, replicas, isProxy = false, memoryBytes = null, inferBatchSize = 1) {
    const data = models[modelType]
    const modelTemplate = data.modelTemplate
    var uri = modelTemplate.uriTemplate
    if (modelTemplate.maxUriSuffix > 0) {
        uri = uri + (uriOffset % modelTemplate.maxUriSuffix).toString()
    }

    const model1 = {"model": {
            "meta":{
                "name": modelName1
            },
            "modelSpec":{
                "uri": uri,
                "requirements": modelTemplate.requirements,
                "memoryBytes": (memoryBytes == null)?modelTemplate.memoryBytes:memoryBytes
            },
            "deploymentSpec": {
                "replicas": replicas
            }
        }
    }

    const model2 = {"model": {
            "meta":{
                "name": modelName2
            },
            "modelSpec":{
                "uri": uri,
                "requirements": modelTemplate.requirements,
                "memoryBytes": (memoryBytes == null)?modelTemplate.memoryBytes:memoryBytes
            },
            "deploymentSpec": {
                "replicas": replicas
            }
        }
    }

    const experiment = {"experiment":{
            "name":experimentName,
            "defaultModel": modelName1,
            "candidates":[
                {"modelName": modelName1,"weight":50},
                {"modelName": modelName2,"weight":50}
            ]
        }
    }

    const inference = getModelInferencePayload(modelType, inferBatchSize)
    return {
        "model1Defn": isProxy ? {"request": model1} : model1,
        "model2Defn": isProxy ? {"request": model2} : model2,
        "experimentDefn": experiment,
        "inference": JSON.parse(JSON.stringify(inference))
    }
}

export function generateModel(modelType, modelName, uriOffset, replicas, isProxy = false, memoryBytes = null, inferBatchSize = 1) {
    const data = models[modelType]
    const modelTemplate = data.modelTemplate
    var uri = modelTemplate.uriTemplate
    if (modelTemplate.maxUriSuffix > 0) {
        uri = uri + (uriOffset % modelTemplate.maxUriSuffix).toString()
    }

    const model = {"model": {
            "meta":{
                "name": modelName
            },
            "modelSpec":{
                "uri": uri,
                "requirements": modelTemplate.requirements,
                "memoryBytes": (memoryBytes == null)?modelTemplate.memoryBytes:memoryBytes
            },
            "deploymentSpec": {
                "replicas": replicas
            }
        }
    }

    // simple one node pipeline
    const pipeline = {"pipeline": {
        "name": generatePipelineName(modelName),
        "steps": [
            {"name": modelName}
        ],
        "output":{
            "steps": [modelName] }
        }
    }

    const inference = getModelInferencePayload(modelType, inferBatchSize)
    return {
        "modelDefn": isProxy ? {"request": model} : model,
        "pipelineDefn": pipeline, // note that we can only deploy a pipeline with a real scheduler
        "inference": JSON.parse(JSON.stringify(inference))
    }
}

export function generatePipelineName(modelName) {
    return modelName + "-pipeline"
}
