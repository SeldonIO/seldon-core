# usage: ./stress-tests.sh [count] [kubectl|seldon] [namespace] [task]
# task is a name of a function that will be called from the list below
# tfsimple
# iris
# tfsimple_pipeline
# tfsimple_join_pipeline

if [ -z "$1" ]
then
      count=5
else
      count=$1
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

if [ -z "$4" ]
then
      task="tfsimple"
else
      task=$4
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

############################################################################################################
# The following functions are the `task` options that can be called by the user.

function tfsimple() {
      echo $i
      sed  's/name: tfsimple1/name: tfsimple'"$1"'/g' models/tfsimple1.yaml > /tmp/models/tfsimple${1}.yaml
      load model /tmp/models/tfsimple${1}.yaml
      status model tfsimple${1}
      seldon model infer tfsimple${1} '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
      seldon model infer tfsimple${1} --inference-mode grpc '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
      unload model tfsimple${1} /tmp/models/tfsimple${1}.yaml
}

function iris() {
      echo $i
      sed  's/name: iris/name: iris'"$1"'/g' models/iris-v1.yaml > /tmp/models/iris${1}.yaml
      load model /tmp/models/iris${1}.yaml
      status model iris${1}
      seldon model infer iris${1} '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
      seldon model infer iris${1} --inference-mode grpc '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
      unload model iris${1} /tmp/models/iris${1}.yaml
}

function tfsimple_pipeline() {
      echo $i
      sed  's/name: tfsimples/name: tfsimples'"$1"'/g' pipelines/tfsimples.yaml > /tmp/pipelines/tfsimples${1}.yaml
      load model ./models/tfsimple1.yaml
      load model ./models/tfsimple2.yaml
      status model tfsimple1
      status model tfsimple2
      load pipeline /tmp/pipelines/tfsimples${1}.yaml
      status pipeline tfsimples${1}
      seldon pipeline infer tfsimples${1} '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
      seldon pipeline infer tfsimples${1} --inference-mode grpc '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
      unload pipeline tfsimples${1} /tmp/pipelines/tfsimples${1}.yaml
      # we cant unload the models here as they are used by other pipelines
      # TODO: create sub models for each pipeline?
      # unload model tfsimple1 ./models/tfsimple1.yaml
      # unload model tfsimple2 ./models/tfsimple2.yaml
}

function tfsimple_join_pipeline() {
      echo $i
      sed  's/name: join/name: join'"$1"'/g' pipelines/tfsimples-join.yaml > /tmp/pipelines/tfsimples-join${1}.yaml
      load model ./models/tfsimple1.yaml
      load model ./models/tfsimple2.yaml
      load model ./models/tfsimple3.yaml
      status model tfsimple1
      status model tfsimple2
      status model tfsimple3
      load pipeline /tmp/pipelines/tfsimples-join${1}.yaml
      status pipeline join${1}
      seldon pipeline infer join${1} '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
      seldon pipeline infer join${1} --inference-mode grpc '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
      unload pipeline join${1} /tmp/pipelines/tfsimples-join${1}.yaml
      # we cant unload the models here as they are used by other pipelines
      # TODO: create sub models for each pipeline?
      # unload model tfsimple1 ./models/tfsimple1.yaml
      # unload model tfsimple2 ./models/tfsimple2.yaml
      # unload model tfsimple3 ./models/tfsimple3.yaml
}

############################################################################################################

mkdir /tmp/models
mkdir /tmp/pipelines
for i in $(seq 1 $count);
do
    $task $i &
done
