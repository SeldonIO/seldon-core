# Development Tips


## Running locally for testing

Sometimes it is useful to be able to test your model locally without the need to build image with s2i or docker.

This can be easily done with `seldon-core` as its installed the CLI command that starts the microservice.

Assuming we have a simple model saved in `MyModel.py` file:
```python
class MyModel:

    def predict(self, X, features_names=None):
        """
        Return a prediction.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        print("Predict called - will run identity function")
        return X
```

We can start Seldon Core microservice with
```bash
seldon-core-microservice MyModel --service-type MODEL
```

Then in other terminal we can send `curl` requests to test REST endpoint:
```bash
curl http://localhost:9000/api/v1.0/predictions \
    -H 'Content-Type: application/json' \
    -d '{"data": {"names": ["input"], "ndarray": ["data"]}}'
```


And assuming that `seldon-core` code is accessible at `${SELDON_CORE_DIR}` we can use `grpcurl` to send gRPC request:
```bash
cd ${SELDON_CORE_DIR}/executor/proto && grpcurl \
    -d '{"data": {"names": ["input"], "ndarray": ["data"]}}' \
    -plaintext -proto ./prediction.proto  0.0.0.0:5000 seldon.protos.Seldon/Predict
```

The `grpcurl` tool can be obtained using binaries released on [GitHub](https://github.com/fullstorydev/grpcurl) or using [asdf-vm](https://github.com/asdf-vm/asdf-plugins).

See [Python Server](./python_server.html#configuration) documentation for config options.
