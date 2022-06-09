# K6 Load Testing

 Available environment variables can be found in `components/settings.js`
 
## Setup

 * Install K6
   * [link](https://k6.io/docs/getting-started/installation/)
 * Local testing
   * `make docker-build-all`
   * `make start-all-mlserver` or `make start-all-triton`

## Examples

 * run tfsimple model
 * output http debug to file log.txt
 * Only load models and infer (no unload)
 * run 500 model loads with 10 concurrent VUs

```
INFER_HTTP_ITERATIONS=10 SKIP_UNLOAD_MODEL=1 MODEL_TYPE="tfsimple" k6 run -u 10 -i 500 --http-debug --log-output=stdout scenarios/load_predict_unload.js > log.txt
```

For isolated agent test (influxdb - grafana)

```
sudo service influxdb start
sudo /bin/systemctl start grafana-server


SCHEDULER_ENDPOINT=0.0.0.0:8100 INFER_GRPC_ENDPOINT=0.0.0.0:8081 INFER_HTTP_ENDPOINT=http://0.0.0.0:8080 INFER_HTTP_ITERATIONS=1 INFER_GRPC_ITERATIONS=1 MODEL_TYPE="iris" MAX_NUM_MODELS=10 k6 run -u 1 -i 10 --http-debug --out influxdb=http://localhost:8086/k6db scenarios/predict_agent.js

SCHEDULER_ENDPOINT=0.0.0.0:8100 INFER_GRPC_ENDPOINT=0.0.0.0:9998 INFER_HTTP_ENDPOINT=http://0.0.0.0:9999 INFER_HTTP_ITERATIONS=1 INFER_GRPC_ITERATIONS=1 MODEL_TYPE="iris" MAX_NUM_MODELS=10 k6 run -u 1 -i 10 --http-debug --out influxdb=http://localhost:8086/k6db scenarios/predict_agent.js
```

For k8s you will need to update the default endpoints to the services exposed, e.g.

```
MODEL_TYPE="tfsimple" SCHEDULER_ENDPOINT=172.18.255.4:9004 INFER_GRPC_ENDPOINT=172.18.255.3:80 INFER_HTTP_ENDPOINT=http://172.18.255.3 k6 run -u 5 -i 50 scenarios/load_predict_unload.js
```

## Constant Throughput Test

Run against model name `iris` which is of type `iris` against a envoy http ip as given.

```
 MODEL_NAME="iris" MODEL_TYPE="iris" INFER_HTTP_ENDPOINT="http://172.31.255.9" k6 run scenarios/model_constant_rate.js
```

Run localy but with grpc

```
INFER_TYPE="grpc" MODEL_TYPE="iris" k6 run scenarios/model_constant_rate.js
```
