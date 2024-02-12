# usage: ./pipeline-tests.sh [sleepTime] [kubectl|seldon] [namespace]

if [ -z "$1" ]
then
      sleepTime=5
else
      sleepTime=$1
fi

if [ -z "$2" ]
then
      cmd="kubectl"
else
      cmd=$2
fi

if [ -z "$3" ]
then
      namespace="seldon-mesh"
else
      namespace=$3
fi

function load() {
  if [ $cmd == "kubectl" ]
  then
      kubectl apply -f $2 -n $namespace
  else
      if [ $1 == "model" ]
      then
            seldon model load -f $2
      elif [ $1 == "pipeline" ]
      then
            seldon pipeline load -f $2
      fi
  fi
}

function unload() {
  if [ $cmd == "kubectl" ]
  then
      kubectl delete -f $3 -n $namespace
  else
      if [ $1 == "model" ]
      then
            seldon model unload $2
      elif [ $1 == "pipeline" ]
      then
            seldon pipeline unload $2
      fi
  fi
}


load model ./models/tfsimple1.yaml
load model ./models/tfsimple2.yaml
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
load "pipeline" ./pipelines/tfsimples.yaml
seldon pipeline status tfsimples -w PipelineReady | jq -M .
seldon pipeline infer tfsimples '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline infer tfsimples --inference-mode grpc '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
unload pipeline tfsimples  ./pipelines/tfsimples.yaml
unload model tfsimple1 ./models/tfsimple1.yaml
unload model tfsimple2 ./models/tfsimple2.yaml

sleep $sleepTime

load model ./models/tfsimple1.yaml
load model ./models/tfsimple2.yaml
load model ./models/tfsimple3.yaml
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon model status tfsimple3 -w ModelAvailable | jq -M .
load pipeline ./pipelines/tfsimples-join.yaml
seldon pipeline status join -w PipelineReady | jq -M .
seldon pipeline infer join --inference-mode grpc     '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
unload pipeline join ./pipelines/tfsimples-join.yaml
unload model tfsimple1 ./models/tfsimple1.yaml
unload model tfsimple2 ./models/tfsimple2.yaml
unload model tfsimple3 ./models/tfsimple3.yaml

sleep $sleepTime

load model ./models/conditional.yaml
load model ./models/add10.yaml
load model ./models/mul10.yaml
seldon model status conditional -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
seldon model status mul10 -w ModelAvailable | jq -M .
load pipeline ./pipelines/conditional.yaml
seldon pipeline status tfsimple-conditional -w PipelineReady | jq -M .
seldon pipeline infer tfsimple-conditional --inference-mode grpc  '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[0]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer tfsimple-conditional --inference-mode grpc  '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
unload pipeline tfsimple-conditional ./pipelines/conditional.yaml
unload model conditional ./models/conditional.yaml
unload model add10 ./models/add10.yaml
unload model mul10 ./models/mul10.yaml

sleep $sleepTime

load model ./models/outlier-error.yaml
seldon model status outlier-error -w ModelAvailable | jq -M .
load pipeline ./pipelines/error.yaml
seldon pipeline status error -w PipelineReady | jq -M .
seldon pipeline infer error --inference-mode grpc     '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer error --inference-mode grpc     '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[100,2,3,4]},"datatype":"FP32","shape":[4]}]}'
unload pipeline error ./models/outlier-error.yaml
unload model outlier-error ./pipelines/error.yaml

sleep $sleepTime

load model ./models/tfsimple1.yaml
load model ./models/tfsimple2.yaml
load model ./models/tfsimple3.yaml
load model ./models/check.yaml
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon model status tfsimple3 -w ModelAvailable | jq -M .
seldon model status check -w ModelAvailable | jq -M .
load pipeline ./pipelines/tfsimples-join-outlier.yaml
seldon pipeline status joincheck -w PipelineReady | jq -M .
seldon pipeline infer joincheck --inference-mode grpc  '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]},"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline infer joincheck --inference-mode grpc     '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
unload pipeline joincheck ./pipelines/tfsimples-join-outlier.yaml
unload model tfsimple1 ./models/tfsimple1.yaml
unload model tfsimple2 ./models/tfsimple2.yaml
unload model tfsimple3 ./models/tfsimple3.yaml
unload model check ./models/check.yaml

sleep $sleepTime

load model ./models/mul10.yaml
load model ./models/add10.yaml
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
load pipeline ./pipelines/pipeline-inputs.yaml
seldon pipeline status pipeline-inputs -w PipelineReady | jq -M .
seldon pipeline infer pipeline-inputs --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
unload pipeline pipeline-inputs ./pipelines/pipeline-inputs.yaml
unload model mul10 ./models/mul10.yaml
unload model add10 ./models/add10.yaml

sleep $sleepTime

load model ./models/mul10.yaml
load model ./models/add10.yaml
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
load pipeline ./pipelines/trigger-joins.yaml
seldon pipeline status trigger-joins -w PipelineReady | jq -M .
seldon pipeline infer trigger-joins --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"ok1","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer trigger-joins --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"ok3","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
unload pipeline trigger-joins ./pipelines/trigger-joins.yaml
unload model mul10 ./models/mul10.yaml
unload model add10 ./models/add10.yaml
