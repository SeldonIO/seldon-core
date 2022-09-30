#!/bin/bash

grpcurl -d '{"model":{
              "meta":{"name":"tfsimple"},
              "modelSpec":{"uri":"gs://seldon-models/triton/simple",
                           "requirements":["tensorflow"],
                           "memoryBytes":500},
              "deploymentSpec":{"replicas":1}}}' \
         -plaintext \
         -import-path ../../../../apis \
         -proto ../../../../apis/mlops/scheduler/scheduler.proto  0.0.0.0:9004 seldon.mlops.scheduler.Scheduler/LoadModel

for i in {0..4000..500}
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
    k6 run ../../scenarios/model_constant_rate.js
done