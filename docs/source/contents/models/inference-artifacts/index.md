# Inference Artifacts

To run your model inside Seldon you must supply an inference artifact that can be downloaded and run on one of MLServer or Triton inference servers. We list artifacts below by alphabetical order below.


```{list-table}
:header-rows: 1

* - Type
  - Server
  - Tag
  - Example
* - Alibi-Detect
  - MLServer
  - `alibi-detect`
  - [example](../../examples/cifar10.md)
* - Alibi-Explain
  - MLServer
  - `alibi-explain`
  - [example](../../examples/explainer-examples.md)
* - DALI
  - Triton
  - `dali`
  - TBC
* - Huggingface
  - MLServer
  - `huggingface`
  - [example](../../examples/huggingface.md)
* - LightGBM
  - MLServer
  - `lightgbm`
  - TBC
* - MLFlow
  - MLServer
  - `mlflow`
  - TBC
* - ONNX
  - Triton
  - `onnx`
  - TBC
* - OpenVino
  - Triton
  - `openvino`
  - TBC
* - Custom Python
  - MLServer
  - `python, mlserver`
  - [example](../../examples/pandasquery.md)
* - Custom Python
  - Triton
  - `python, triton`
  - TBC
* - PyTorch
  - Triton
  - `pytorch`
  - TBC
* - SKLearn
  - MLServer
  - `sklearn`
  - [example](../../examples/income.md)
* - Spark Mlib
  - MLServer
  - `spark-mlib`
  - TBC
* - Tensorflow
  - Triton
  - `tensorflow`
  - [example](../../examples/cifar10.md)
* - TensorRT
  - Triton
  - `tensorrt`
  - TBC
* - Triton FIL
  - Triton
  - `fil`
  - TBC
* - XGBoost
  - MLServer
  - `xgboost`
  - TBC
```

## Saving Model artifacts

For many machine learning artifacts you can simply save them to a folder and load them into seldon core v2. Details are given below as well as a link to creating a custom model settings file if needed.

```{list-table}
:header-rows: 1

* - Type
  - Notes
  - Custom Model Settings

* - Alibi-Detect
  - [Save model using Alibi-Detect](https://docs.seldon.io/projects/alibi-detect/en/stable/overview/saving.html)
  - [docs](https://docs.seldon.io/projects/alibi-detect/en/stable/)
* - Alibi-Explain
  - [Save model using Alibi-Explain](https://docs.seldon.io/projects/alibi/en/stable/overview/saving.html)
  - [docs](https://docs.seldon.io/projects/alibi/en/stable/)
* - DALI
  - 
  - [docs](https://github.com/triton-inference-server/dali_backend)
* - Huggingface
  - Create an MLServer model-settings.json with the Huggingface model required
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/huggingface/README.md)
* - LightGBM
  - Save model to file with extension`.bst`
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/lightgbm/README.md)
* - MLFlow
  - 
  - [docs](https://github.com/SeldonIO/MLServer/tree/master/runtimes/mlflow)
* - ONNX
  - 
  - [docs](https://github.com/triton-inference-server/onnxruntime_backend)
* - OpenVino
  - 
  - [docs](https://github.com/triton-inference-server/openvino_backend)
* - Custom MLServer Python
  - Create a python file with a class that extends `MLModel`.
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/custom/README.md)
* - Custom Triton Python
  - 
  - [docs](https://github.com/triton-inference-server/python_backend)
* - PyTorch
  - 
  - [docs](https://github.com/triton-inference-server/pytorch_backend)
* - SKLearn
  - Save model via joblib to a file with extension `.joblib` or with pickle to a file with extension `.pkl` or `.pickle`
  - [docs](https://github.com/SeldonIO/MLServer/tree/master/runtimes/sklearn)
* - Spark Mlib
  - 
  - [docs](https://github.com/SeldonIO/MLServer/tree/master/runtimes/mllib)
* - Tensorflow
  - Save model as Saved Model format. If using graphdef format you will need to create Triton config.pbtxt.
  - [docs](https://github.com/triton-inference-server/tensorflow_backend)
* - TensorRT
  - 
  - [docs](https://github.com/triton-inference-server/tensorrt_backend)
* - Triton FIL
  - 
  - [docs](https://github.com/triton-inference-server/fil_backend)
* - XGBoost
  - Save model to file with extension`.bst` or `.json`
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/xgboost/README.md)
```

## Custom MLServer Model Settings

For [MLServer](https://github.com/SeldonIO/MLServer) targeted models you can create a model-settings.json file to help MLServer load your model and place this alongside your artifact. See the [MLServer project](https://mlserver.readthedocs.io/en/latest/reference/model-settings.html)  for details.

## Custom Triton Configuration

For [Triton inference server](https://github.com/triton-inference-server/server) models you can create [a configuration config.pbtxt file](https://github.com/triton-inference-server/server/blob/main/docs/user_guide/model_configuration.md) alongside your artifact.

## Notes

 * The `tag` field represents the tag you need to add to the `requirements` part of the Model spec for your artifact to be loaded on a compatible server. e.g. for an sklearn model:
   ```{literalinclude} ../../../../../samples/models/sklearn-iris-gs.yaml 
   :language: yaml
   ```


