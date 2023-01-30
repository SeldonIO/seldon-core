#!/bin/bash

NUM_MODELS="1,100"
MODEL_NAME="mlflow_wine,tfsimple" # check model.js
MODELNAME_PREFIX_VAR="mlflow_winea,tfsimplea"
INFER_BATCH_SIZE_VAR="1,1"
MODEL_MEMORY_BYTES_VAR="400,500"

NUM_ITERS=500
NUM_VUS=5
EXTRA="--out influxdb=http://localhost:8086/k6db ../../scenarios/infer_constant_vu.js"
DIR="results"

mkdir -p $DIR


export SCHEDULER_ENDPOINT=0.0.0.0:9004 
export INFER_GRPC_ENDPOINT=0.0.0.0:9000
export INFER_HTTP_ENDPOINT=http://0.0.0.0:9000
export INFER_HTTP_ITERATIONS=1 
export INFER_GRPC_ITERATIONS=1 
export MODEL_TYPE=$MODEL_NAME
export MAX_NUM_MODELS=$NUM_MODELS
export SCHEDULER_PROXY="false"
export ENVOY="true"
export INFER_BATCH_SIZE=$INFER_BATCH_SIZE_VAR
export MODELNAME_PREFIX=$MODELNAME_PREFIX_VAR
export MODEL_MEMORY_BYTES=$MODEL_MEMORY_BYTES_VAR
#export DATAFLOW_TAG="pipeline"  # "model" or "pipeline" or "", "pipeline" would trigger kstreams
#export SKIP_UNLOAD_MODEL=1

k6 run --http-debug="full" -u $NUM_VUS -i $NUM_ITERS --summary-export $DIR/base.json --out csv=$DIR/base.gz $EXTRA


#export INFER_GRPC_ENDPOINT=0.0.0.0:9998 
#export INFER_HTTP_ENDPOINT=http://0.0.0.0:9999 


#k6 run -u $NUM_VUS -i $NUM_ITERS --summary-export $DIR/oc_$MAX_NUM_MODELS.json --out csv=$DIR/$MAX_NUM_MODELS.gz $EXTRA

#export MAX_NUM_MODELS=$(( $NUM_MODELS + 5 ))
#k6 run -u $NUM_VUS -i $NUM_ITERS --summary-export $DIR/oc_$MAX_NUM_MODELS.json --out csv=$DIR/$MAX_NUM_MODELS.gz $EXTRA

#export MAX_NUM_MODELS=$(( $NUM_MODELS + 10 ))
#k6 run -u $NUM_VUS -i $NUM_ITERS --summary-export $DIR/oc_$MAX_NUM_MODELS.json --out csv=$DIR/$MAX_NUM_MODELS.gz $EXTRA

#export MAX_NUM_MODELS=$(( $NUM_MODELS + 25 ))
#k6 run -u $NUM_VUS -i $NUM_ITERS --summary-export $DIR/oc_$MAX_NUM_MODELS.json --out csv=$DIR/$MAX_NUM_MODELS.gz $EXTRA