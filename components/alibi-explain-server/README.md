# Alibi Model Explainer

[Alibi](https://github.com/SeldonIO/alibi) server is an implementation for providing black box model explanation for KFServer models.

To start the server locally for development needs, run the following command under this folder in your github repository. 

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
The first command will use `gsutil` to download an SKLearn Iris model from Seldon's public GCS.
The second command will create an Anchor Tabular explainer model using utils in the package.

Then, launch an SKLearn server with the model that was downloaded in the previous command, within Docker:
```bash
make anchor_tabular_model
```

Verify that it's correctly able to handle a prediction request:
```bash
make anchor_tabular_predict
```

And, launch the local Anchor Explainer:
```bash
make anchor_tabular
```

Verify that the explainer can do an explanation:
```bash
make anchor_tabular_explain
```

Finally, to run the explainer in a Dockerfile instead of a stand-alone program:
```bash
make anchor_tabular_docker
```

And, you can verify a successful explanation again with:
```bash
make anchor_tabular_explain
```

#### Tests for Text Explanations

To ensure you have the necessary model artefact:
```bash
make test_models/sklearn/moviesentiment
```

Launch the SKLearn server with the model loaded:
```bash
make anchor_text_model
```

Send a prediction request:
```bash
make anchor_text_predict
```

Next, launch a text explainer:
```bash
make anchor_text
```

Send a request to the explainer:
```bash
make anchor_text_explain
```

Finally, to run the explainer in a Dockerfile instead of a stand-alone program:
```bash
make anchor_text_docker
```

And, you can verify a successful explanation again with:
```bash
make anchor_text_explain
```

#### Tests for Image Explanations

To ensure you have the necessary model artefact:
```bash
make test_models/tfserving/cifar10/resnet32
```

```bash
make test_models/explainers/anchor_image
```

```bash
make anchor_images_model
```

```bash
make anchor_images_predict
```

```bash
make anchor_images
```

```bash
make anchor_images_explain
```

```bash
make anchor_images_docker
```

```bash
make anchor_images_explain
```


### Notes

 * Pythin 3.8 ubi can not be presently used because this would require all explainer models to be saved as python 3.8 in examples otherwise dill load issues happen.
 * KernelShap has issues for models not saved in matching python to Docker container - e.g., py36
