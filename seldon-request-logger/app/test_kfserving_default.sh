
# now handling batch and inferring type in the cifar10 logger
# will just replace default logger with it
# need to check all the scenarios

# now handling tftensor https://github.com/SeldonIO/seldon-core/pull/1669/files#diff-5df915d0fb8c5f6187e4f45ec2f364d6R59
# also handling request.names like request.instance but just for names
curl 0.0.0.0:2222 -d '{"instances": [[6.8,  2.8,  4.8,  1.4],[6.0,  3.4,  4.5,  1.6]]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.request" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: iris" -H "Ce-Requestid: 8h"
curl 0.0.0.0:2222 -d '{"predictions": [1,1]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.response" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: iris" -H "Ce-Requestid: 8h"
curl 0.0.0.0:2222 -d '{"data": {"feature_score": null, "instance_score": null, "is_outlier": [1, 1]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.outlier" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: iris" -H "Ce-id: 8h"