seldon model load -f ./models/tfsimple1.yaml 
seldon model load -f ./models/tfsimple2.yaml 
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/tfsimples.yaml
seldon pipeline status tfsimples -w PipelineReady| jq -M .
seldon pipeline infer tfsimples '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline infer tfsimples --inference-mode grpc '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline unload tfsimples
seldon model unload tfsimple1
seldon model unload tfsimple2


seldon model load -f ./models/tfsimple1.yaml 
seldon model load -f ./models/tfsimple2.yaml 
seldon model load -f ./models/tfsimple3.yaml 
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon model status tfsimple3 -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/tfsimples-join.yaml
seldon pipeline status join -w PipelineReady | jq -M .
seldon pipeline infer join --inference-mode grpc     '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline unload join
seldon model unload tfsimple1
seldon model unload tfsimple2
seldon model unload tfsimple3

seldon model load -f ./models/conditional.yaml 
seldon model load -f ./models/add10.yaml 
seldon model load -f ./models/mul10.yaml 
seldon model status conditional -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
seldon model status mul10 -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/conditional.yaml
seldon pipeline status tfsimple-conditional -w PipelineReady | jq -M .
seldon pipeline infer tfsimple-conditional --inference-mode grpc  '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[0]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer tfsimple-conditional --inference-mode grpc  '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline unload tfsimple-conditional
seldon model unload conditional
seldon model unload add10
seldon model unload mul10


seldon model load -f ./models/outlier-error.yaml 
seldon model status outlier-error -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/error.yaml
seldon pipeline status error -w PipelineReady | jq -M .
seldon pipeline infer error --inference-mode grpc     '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer error --inference-mode grpc     '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[100,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
seldon pipeline unload error
seldon model unload outlier-error


seldon model load -f ./models/tfsimple1.yaml 
seldon model load -f ./models/tfsimple2.yaml 
seldon model load -f ./models/tfsimple3.yaml 
seldon model load -f ./models/check.yaml 
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon model status tfsimple3 -w ModelAvailable | jq -M .
seldon model status check -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/tfsimples-join-outlier.yaml
seldon pipeline status joincheck -w PipelineReady | jq -M .
seldon pipeline infer joincheck --inference-mode grpc  '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]},"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline infer joincheck --inference-mode grpc     '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
seldon pipeline unload joincheck
seldon model unload tfsimple1
seldon model unload tfsimple2
seldon model unload tfsimple3
seldon model unload check


seldon model load -f ./models/mul10.yaml 
seldon model load -f ./models/add10.yaml 
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/pipeline-inputs.yaml
seldon pipeline status pipeline-inputs -w PipelineReady | jq -M .
seldon pipeline infer pipeline-inputs --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
seldon pipeline unload pipeline-inputs
seldon model unload mul10
seldon model unload add10


seldon model load -f ./models/mul10.yaml 
seldon model load -f ./models/add10.yaml 
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/trigger-joins.yaml
seldon pipeline status trigger-joins -w PipelineReady | jq -M .
seldon pipeline infer trigger-joins --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"ok1","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline infer trigger-joins --inference-mode grpc  '{"model_name":"pipeline","inputs":[{"name":"ok3","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
seldon pipeline unload trigger-joins
seldon model unload mul10
seldon model unload add10
