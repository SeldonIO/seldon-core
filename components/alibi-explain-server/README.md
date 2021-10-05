# Alibi Model Explainer

[Alibi](https://github.com/SeldonIO/alibi) server is an implementation for providing black box model explanation for KFServer models.

To start the server locally for development needs, run the following command under this folder in your github repository. 

This server uses [Poetry](https://python-poetry.org/docs/) to manage its environment.
Please make sure you install poetry before continuing.

## Local development

### Python environment

1. If you have `asdf-vm` installed the `.tool-versions` will specify required version of Python
2. Otherwise create and activate conda environment with appropriate Python version

### Install dependencies

To install dependencies run
```bash
poetry install
```

Alternativey use the Makefile target:
```bash
make dev_install
```

### Running Tests

To run tests run
```bash
poetry run pytest -v .
```

Alternatively use the Makefile target:
```bash
make test
```

The Makefile also has tests for full worked examples.


### Notes

 * Pythin 3.8 ubi can not be presently used because this would require all explainer models to be saved as python 3.8 in examples otherwise dill load issues happen.
 * KernelShap has issues for models not saved in matching python to Docker container - e.g., py36
