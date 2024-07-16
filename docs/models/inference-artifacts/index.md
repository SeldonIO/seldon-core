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
  - [example](../../examples/model-zoo.md#lightgbm-model)
* - MLFlow
  - MLServer
  - `mlflow`
  - [example](../../examples/model-zoo.md#mlflow-wine-model)
* - ONNX
  - Triton
  - `onnx`
  - [example](../../examples/model-zoo.md#onnx-mnist-model)
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
  - [example](https://github.com/SeldonIO/triton-python-examples)
* - PyTorch
  - Triton
  - `pytorch`
  - [example](../../examples/model-zoo.md#pytorch-mnist-model)
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
  - [example](../../examples/model-zoo.md#xgboost-model)
```

## Saving Model artifacts

For many machine learning artifacts you can simply save them to a folder and load them into seldon core v2. Details are given below as well as a link to creating a custom model settings file if needed.

```{list-table}
:header-rows: 1

* - Type
  - Notes
  - Custom Model Settings

* - Alibi-Detect
  - [Save model using Alibi-Detect](https://docs.seldon.io/projects/alibi-detect/en/stable/overview/saving.html).
  - [docs](https://docs.seldon.io/projects/alibi-detect/en/stable/)
* - Alibi-Explain
  - [Save model using Alibi-Explain](https://docs.seldon.io/projects/alibi/en/stable/overview/saving.html).
  - [docs](https://docs.seldon.io/projects/alibi/en/stable/)
* - DALI
  - Follow the Triton docs to create a config.pbtxt and model folder with artifact.
  - [docs](https://github.com/triton-inference-server/dali_backend)
* - Huggingface
  - Create an MLServer `model-settings.json` with the Huggingface model required
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/huggingface/README.md)
* - LightGBM
  - Save model to file with extension`.bst`.
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/lightgbm/README.md)
* - MLFlow
  - Use the created `artifacts/model` folder from your training run.
  - [docs](https://github.com/SeldonIO/MLServer/tree/master/runtimes/mlflow)
* - ONNX
  - Save you model with name `model.onnx`.
  - [docs](https://github.com/triton-inference-server/onnxruntime_backend)
* - OpenVino
  - Follow the Triton docs to create your model artifacts.
  - [docs](https://github.com/triton-inference-server/openvino_backend)
* - Custom MLServer Python
  - Create a python file with a class that extends `MLModel`.
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/custom/README.md)
* - Custom Triton Python
  - Follow the Triton docs to create your `config.pbtxt` and associated python files.
  - [docs](https://github.com/triton-inference-server/python_backend)
* - PyTorch
  - Create a Triton `config.pbtxt` describing inputs and outputs and place traced torchscript in folder as `model.pt`.
  - [docs](https://github.com/triton-inference-server/pytorch_backend)
* - SKLearn
  - Save model via joblib to a file with extension `.joblib` or with pickle to a file with extension `.pkl` or `.pickle`.
  - [docs](https://github.com/SeldonIO/MLServer/tree/master/runtimes/sklearn)
* - Spark Mlib
  - Follow the MLServer docs.
  - [docs](https://github.com/SeldonIO/MLServer/tree/master/runtimes/mllib)
* - Tensorflow
  - Save model in "Saved Model" format as `model.savedodel`. If using graphdef format you will need to create Triton config.pbtxt and place your model in a numbered sub folder. HDF5 is not supported.
  - [docs](https://github.com/triton-inference-server/tensorflow_backend)
* - TensorRT
  - Follow the Triton docs to create your model artifacts.
  - [docs](https://github.com/triton-inference-server/tensorrt_backend)
* - Triton FIL
  - Follow the Triton docs to create your model artifacts.
  - [docs](https://github.com/triton-inference-server/fil_backend)
* - XGBoost
  - Save model to file with extension`.bst` or `.json`.
  - [docs](https://github.com/SeldonIO/MLServer/blob/master/docs/examples/xgboost/README.md)
```

## Custom MLServer Model Settings

For [MLServer](https://github.com/SeldonIO/MLServer) targeted models you can create a `model-settings.json` file to help MLServer load your model and place this alongside your artifact. See the [MLServer project](https://mlserver.readthedocs.io/en/latest/reference/model-settings.html)  for details.

## Custom Triton Configuration

For [Triton inference server](https://github.com/triton-inference-server/server) models you can create [a configuration config.pbtxt file](https://github.com/triton-inference-server/server/blob/main/docs/user_guide/model_configuration.md) alongside your artifact.

## Notes

 * The `tag` field represents the tag you need to add to the `requirements` part of the Model spec for your artifact to be loaded on a compatible server. e.g. for an sklearn model:
   ```{literalinclude} ../../../../../samples/models/sklearn-iris-gs.yaml 
   :language: yaml
   ```


