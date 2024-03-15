# usage: ./smoke-tests.sh [sleepTime] [kubectl|seldon] [namespace]

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
      elif [ $1 == "experiment" ]
      then
            seldon experiment start -f $2
      fi
  fi
}

function unload() {
  # note that in k8s we need to use the filepath (3rd argument)
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
      elif [ $1 == "experiment" ]
      then
            seldon experiment stop $2
      fi
  fi
}

function status() {
  if [ $cmd == "kubectl" ]
  then
      if [ $1 == "model" ]
      then
            kubectl wait --for condition=ready --timeout=300s model $2 -n $namespace
      elif [ $1 == "pipeline" ]
      then
            kubectl wait --for condition=ready --timeout=300s pipeline $2 -n $namespace
      elif [ $1 == "experiment" ]
      then
            kubectl wait --for condition=ready --timeout=300s experiment $2 -n $namespace
      fi
  else
      if [ $1 == "model" ]
      then
            seldon model status $2 -w ModelAvailable | jq -M .
      elif [ $1 == "pipeline" ]
      then
            seldon pipeline status $2 -w PipelineReady | jq -M .
      elif [ $1 == "experiment" ]
      then
            seldon experiment status $2 -w | jq -M .
      fi
  fi
}


load model ./models/tfsimple1.yaml
load model ./models/tfsimple2.yaml
status model tfsimple1
status model tfsimple2 
load "pipeline" ./pipelines/tfsimples.yaml
status pipeline tfsimples
seldon pipeline infer tfsimples '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline infer tfsimples --inference-mode grpc '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
unload pipeline tfsimples  ./pipelines/tfsimples.yaml
unload model tfsimple1 ./models/tfsimple1.yaml
unload model tfsimple2 ./models/tfsimple2.yaml

sleep $sleepTime

load model ./models/tfsimple1.yaml
load model ./models/tfsimple2.yaml
load model ./models/tfsimple3.yaml
status model tfsimple1
status model tfsimple2
status model tfsimple3
load pipeline ./pipelines/tfsimples-join.yaml
status pipeline join
seldon pipeline infer join --inference-mode grpc     '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
unload pipeline join ./pipelines/tfsimples-join.yaml
unload model tfsimple1 ./models/tfsimple1.yaml
unload model tfsimple2 ./models/tfsimple2.yaml
unload model tfsimple3 ./models/tfsimple3.yaml

sleep $sleepTime

load model ./models/conditional.yaml
load model ./models/add10.yaml
load model ./models/mul10.yaml
status model conditional
status model add10
status model mul10
load pipeline ./pipelines/conditional.yaml
status pipeline tfsimple-conditional
seldon pipeline infer tfsimple-conditional --inference-mode grpc  '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[0]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer tfsimple-conditional --inference-mode grpc  '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
unload pipeline tfsimple-conditional ./pipelines/conditional.yaml
unload model conditional ./models/conditional.yaml
unload model add10 ./models/add10.yaml
unload model mul10 ./models/mul10.yaml

sleep $sleepTime

load model ./models/outlier-error.yaml
status model outlier-error
load pipeline ./pipelines/error.yaml
status pipeline error
seldon pipeline infer error --inference-mode grpc     '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer error --inference-mode grpc     '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[100,2,3,4]},"datatype":"FP32","shape":[4]}]}'
unload pipeline error ./models/outlier-error.yaml
unload model outlier-error ./pipelines/error.yaml

sleep $sleepTime

load model ./models/tfsimple1.yaml
load model ./models/tfsimple2.yaml
load model ./models/tfsimple3.yaml
load model ./models/check.yaml
status model tfsimple1
status model tfsimple2
status model tfsimple3
status model check
load pipeline ./pipelines/tfsimples-join-outlier.yaml
status pipeline joincheck
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
status model mul10
status model add10
load pipeline ./pipelines/pipeline-inputs.yaml
status pipeline pipeline-inputs
seldon pipeline infer pipeline-inputs --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
unload pipeline pipeline-inputs ./pipelines/pipeline-inputs.yaml
unload model mul10 ./models/mul10.yaml
unload model add10 ./models/add10.yaml

sleep $sleepTime

load model ./models/mul10.yaml
load model ./models/add10.yaml
status model mul10
status model add10
load pipeline ./pipelines/trigger-joins.yaml
status pipeline trigger-joins
seldon pipeline infer trigger-joins --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"ok1","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer trigger-joins --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"ok3","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
unload pipeline trigger-joins ./pipelines/trigger-joins.yaml
unload model mul10 ./models/mul10.yaml
unload model add10 ./models/add10.yaml


# MLServer
sleep $sleepTime
load model ./models/sklearn-iris-gs.yaml
status model iris
seldon model infer iris '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
seldon model infer iris --inference-mode grpc \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
unload model iris ./models/sklearn-iris-gs.yaml


# Experiments
load model ./models/sklearn1.yaml
load model ./models/sklearn2.yaml
status model iris
status model iris2
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
load experiment ./experiments/ab-default-model.yaml 
status experiment experiment-sample
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
seldon model infer iris --show-headers \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
seldon model infer iris -s -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
unload experiment experiment-sample ./experiments/ab-default-model.yaml
unload model iris ./models/sklearn1.yaml
unload model iris2 ./models/sklearn2.yaml
