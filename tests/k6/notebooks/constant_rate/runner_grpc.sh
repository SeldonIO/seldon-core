#!/bin/bash

SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT:-0.0.0.0:9004}"
INFER_GRPC_ENDPOINT="${INFER_GRPC_ENDPOINT:-0.0.0.0:9000}"
INFER_HTTP_ENDPOINT="${INFER_HTTP_ENDPOINT:-http://0.0.0.0:9000}"

for i in {500..2000..500}
do
    if [ $i -eq 0 ]; 
    then
        ii=1
    else
        ii=$i
    fi
    echo "Request rate: "$ii
    MODEL_NAME=iris \
    MODEL_TYPE=iris \
    INFER_TYPE=GRPC \
    INFER_BATCH_SIZE=1 \
    REQUEST_RATE=$ii \
    SCHEDULER_ENDPOINT=$SCHEDULER_ENDPOINT \
    INFER_GRPC_ENDPOINT=$INFER_GRPC_ENDPOINT \
    INFER_HTTP_ENDPOINT=$INFER_HTTP_ENDPOINT \
    MAX_NUM_MODELS=1 \
    k6 run ../../scenarios/infer_constant_rate.js
done
