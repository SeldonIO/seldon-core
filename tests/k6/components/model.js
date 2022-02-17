
const models = {
    "iris": {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/testing/iris",
            "maxUriSuffix": 10,
            "requirements": ["sklearn"],
            "memoryBytes": 20000,
        },
        "inference":{
            "http": {"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]},
            "grpc": {"inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}
        }
    },
    "tfsimple": {
        "modelTemplate": {
            "uriTemplate": "gs://seldon-models/triton/simple",
            "maxUriSuffix": 0,
            "requirements": ["tensorflow"],
            "memoryBytes": 20000,
        },
        "inference":{
            "http": {"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]},
            "grpc": {"inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}
        }
    },
}

export function generateModel(modelType, modelName, uriOffset, replicas, isProxy = false) {
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
                "memoryBytes": modelTemplate.memoryBytes
            },
            "deploymentSpec": {
                "replicas": replicas
            }
        }
    }
    //console.log(JSON.stringify(model))


    const inference = data.inference
    return {
        "modelDefn": isProxy ? {"request": model} : model,
        "inference": JSON.parse(JSON.stringify(inference))
    }
}