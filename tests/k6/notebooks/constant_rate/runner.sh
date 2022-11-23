#!/bin/bash

SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT:-0.0.0.0:9004}"
INFER_GRPC_ENDPOINT="${INFER_GRPC_ENDPOINT:-0.0.0.0:9000}"
INFER_HTTP_ENDPOINT="${INFER_HTTP_ENDPOINT:-http://0.0.0.0:9000}"


grpcurl -d '{"model":{
              "meta":{"name":"tfsimple"},
              "modelSpec":{"uri":"gs://seldon-models/triton/simple",
                           "requirements":["tensorflow"],
                           "memoryBytes":500},
              "deploymentSpec":{"minReplicas":1, "replicas":1}}}' \
         -plaintext \
         -import-path ../../../../apis \
         -proto ../../../../apis/mlops/scheduler/scheduler.proto  $SCHEDULER_ENDPOINT seldon.mlops.scheduler.Scheduler/LoadModel

for i in {0..4000..200}
do
    if [ $i -eq 0 ]; 
    then
        ii=1
    else
        ii=$i
    fi
    echo "Request rate: "$ii
    MODEL_NAME=tfsimple \
    MODEL_TYPE=tfsimple \
    INFER_TYPE=REST \
    INFER_BATCH_SIZE=1 \
    REQUEST_RATE=$ii \
    SCHEDULER_ENDPOINT=$SCHEDULER_ENDPOINT \
    INFER_GRPC_ENDPOINT=$INFER_GRPC_ENDPOINT \
    INFER_HTTP_ENDPOINT=$INFER_HTTP_ENDPOINT \
    k6 run ../../scenarios/model_constant_rate.js
done

grpcurl -d '{"model":{"name":"tfsimple"}}' \
         -plaintext \
         -import-path ../../../../apis \
         -proto ../../../../apis/mlops/scheduler/scheduler.proto  $SCHEDULER_ENDPOINT seldon.mlops.scheduler.Scheduler/UnloadModel
