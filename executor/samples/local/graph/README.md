# Inference Graph Tests

## Transformer and Model

Ensure executor is built

```
cd ../../.. && make executor
```

In different terminals:

Run executor

```
make run_rest_executor
```

Run transformer

```
make run_transformer
```

Run model

```
make run_model
```

Test multipart curl

```
make curl_rest_multipart
```

Test normal curl

```
make curl_rest
```