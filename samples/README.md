# Samples

Various example notebooks.

## Requirements

 * a seldon cli
 * jupyter notebook
 * jq

## Artefacts

All (Python) artefacts are versioned by their respective version of Triton and
MLServer used at the time of training (e.g.
`gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn`).
To re-generate them, you can call the following Makefile targets:

```
$ make train-all upload-all
```

The new version of the artefacts will get uploaded to GCS, using the version of
Triton and MLServer specified on [scheduler/Makefile](../scheduler/Makefile).
The `upload-all` target will also replace all previous references to versioned
artefacts in the notebooks.
 
