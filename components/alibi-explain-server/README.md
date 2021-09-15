# Alibi Model Explainer

[Alibi](https://github.com/SeldonIO/alibi) server is an implementation for providing black box model explanation for KFServer models.

To start the server locally for development needs, run the following command under this folder in your github repository. 

```
make dev_install
```

Run tests

```
make test
```

The Makefile also has tests for full worked examples.

## Development

Install the development dependencies with:

```bash
make dev_install
```

### Notes

 * Pythin 3.8 ubi can not be presently used because this would require all explainer models to be saved as python 3.8 in examples otherwise dill load issues happen.
 * KernelShap has issues for models not saved in matching python to Docker container - e.g., py36


