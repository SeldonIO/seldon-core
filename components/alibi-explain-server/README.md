# Alibi Model Explainer

[Alibi](https://github.com/SeldonIO/alibi) server is an implementation for providing black box model explanation for KFServer models.

To start the server locally for development needs, run the following command under this folder in your GitHub repository.

This server uses [Poetry](https://python-poetry.org/docs/) to manage its environment.
Please make sure you install poetry before continuing.

## Local development

### Python environment

1. If you have `asdf-vm` installed the `.tool-versions` will specify required version of Python
2. Otherwise, create and activate conda environment with appropriate Python version

### Install dependencies

To install dependencies run:
```bash
poetry install
```

Alternative use the Makefile target:
```bash
make dev_install
```

Then, you will have to generate Python files for the protos:
```bash
make build_apis
```

### Running Unit Tests

To run tests run
```bash
poetry run pytest -v .
```

Alternatively use the Makefile target:
```bash
make test
```

The Makefile also has tests for full worked examples.


### Running Integration Tests

#### Tests for Tabular Explanations

First, to ensure you have the necessary artefacts for the tests:
```bash
make test_models/sklearn/iris
make test_models/explainers/anchor_tabular
```

To run the SKLearn model in a Docker container and send a prediction request:
```bash
make anchor_tabular_model
make anchor_tabular_predict
```

##### Send an explanation request with the explainer running as a stand-alone program
* You need the SKLearn model running in Docker
* You need the explainer running locally
```bash
make anchor_tabular_model
make anchor_tabular
make anchor_tabular_explain
```

##### Send an explanation request with the explainer running in Docker
* You need the SKLearn model running in Docker
* You need the explainer running in Docker
```bash
make anchor_tabular_model
make anchor_tabular_docker
make anchor_tabular_explain
```

#### Tests for Text Explanations

To ensure you have the necessary model artefact:
```bash
make test_models/sklearn/moviesentiment
```

To run the SKLearn model in a Docker container and send a prediction request:
```bash
make anchor_text_model
make anchor_text_predict
```

##### Send an explanation request with the explainer running as a stand-alone program
* You need the SKLearn model running in Docker
* You need the explainer running locally
```bash
make anchor_text_model
make anchor_text
make anchor_text_explain
```

##### Send an explanation request with the explainer running in Docker
* You need the SKLearn model running in Docker
* You need the explainer running in Docker
```bash
make anchor_text_model
make anchor_text_docker
make anchor_text_explain
```

#### Tests for Image Explanations

First, to ensure you have the necessary artefacts for the tests:
```bash
make test_models/tfserving/cifar10/resnet32
make test_models/explainers/anchor_image
```

To run the TF model in a Docker container and send a prediction request:
```bash
make anchor_images_model
make anchor_images_predict
```

##### Send an explanation request with the explainer running as a stand-alone program
* You need the TF model running in Docker
* You need the explainer running locally
```bash
make anchor_images_model
make anchor_images
make anchor_images_explain
```

##### Send an explanation request with the explainer running in Docker
* You need the TF model running in Docker
* You need the explainer running in Docker
```bash
make anchor_images_model
make anchor_images_docker
make anchor_images_explain
```

#### Tests for Kernel Shap Explanations

First, to ensure you have the necessary artefacts for the tests:
```bash
make test_models/sklearn/wine/model-py36-0.23.2
make test_models/explainers/kernel_shap
```

To run the SKLearn model in a Docker container and send a prediction request:
```bash
make kernel_shap_model
make kernel_shap_predict
```

##### Send an explanation request with the explainer running as a stand-alone program
* You need the SKLearn model running in Docker
* You need the explainer running locally
```bash
make kernel_shap_model
make kernel_shap
make kernel_shap_explain
```

##### Send an explanation request with the explainer running in Docker
* You need the SKLearn model running in Docker
* You need the explainer running in Docker
```bash
make kernel_shap_model
make kernel_shap_docker
make kernel_shap_explain
```

#### Tests for Integrated Gradients

To ensure you have the necessary model artefact:
```bash
make test_models/keras/imdb
```

##### Send an explanation request with the explainer running as a stand-alone program
* You need the Keras explainer running locally
```bash
make integrated_gradients
make integrated_gradients_explain
```

##### Send an explanation request with the explainer running in Docker
* You need the Keras explainer running in Docker
```bash
make integrated_gradients_docker
make integrated_gradients_explain
```

#### Tests for Tree Shap

To ensure you have the necessary model artefact:
```bash
make test_models/explainers/tree_shap
```

##### Send an explanation request with the explainer running as a stand-alone program
* You need the explainer running locally
```bash
make tree_shap
make tree_shap_explain
```

##### Send an explanation request with the explainer running in Docker
* You need the explainer running in Docker
```bash
make tree_shap_docker
make tree_shap_explain
```

#### Tests for ALE

First, to ensure you have the necessary artefacts for the tests:
```bash
make test_models/sklearn/iris-0.23.2/lr_model
make test_models/explainers/ale
```

To run the SKLearn model in a Docker container and send a prediction request:
```bash
make ale_model
make ale_predict
```

##### Send an explanation request with the explainer running as a stand-alone program
* You need the explainer running locally
```bash
make ale_model
make ale
make ale_explain
```

##### Send an explanation request with the explainer running in Docker
* You need the explainer running in Docker
```bash
make ale_model
make ale_docker
make ale_explain
```

### Notes

* Pythin 3.8 ubi can not be presently used because this would require all explainer models to be saved as python 3.8 in examples otherwise dill load issues happen.
* KernelShap has issues for models not saved in matching python to Docker container - e.g., py36
