# K6 Load Testing

 Available environment variables can be found in `components/settings.js`
 
## Setup

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

For k8s you will need to update the default endpoints to the services exposed, e.g.

```
MODEL_TYPE="tfsimple" SCHEDULER_ENDPOINT=172.18.255.4:9004 INFER_GRPC_ENDPOINT=172.18.255.3:80 INFER_HTTP_ENDPOINT=http://172.18.255.3 k6 run -u 5 -i 50 scenarios/load_predict_unload.js
```