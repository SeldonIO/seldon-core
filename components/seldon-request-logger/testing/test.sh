#start local logger and elastic to run - see README

#seldon tensor
curl 0.0.0.0:2222 -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: tensor" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 1a"
curl 0.0.0.0:2222 -d '{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: tensor" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Requestid: 1a"

#batch seldon ndarray
curl 0.0.0.0:2222 -d '{"data":{"names":["a","b"],"ndarray":[[1,2],[3,4]]}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: ndarray" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 2b"
curl 0.0.0.0:2222 -d '{"data":{"names":["c"],"ndarray":[[7],[8]]}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: ndarray" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Requestid: 2b"

#ndarray containing strings (batch) - also using modelid but not inferenceservicename
curl 0.0.0.0:2222 -d '{"data":{"names":["a"],"ndarray":["test1","test2"]}}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 3c" \
  -H "Ce-Modelid: classifier" -H "Ce-Namespace: default" -H "Ce-Endpoint: example"
curl 0.0.0.0:2222 -d '{"data":{"names":["c"],"ndarray":[[7],[8]]}}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Requestid: 3c" \
  -H "Ce-Modelid: classifier" -H "Ce-Namespace: default" -H "Ce-Endpoint: example"

#tensor again (two batches of tabular) but this time no inferenceservice name or modelid - still allowed but will go to unknown
curl 0.0.0.0:2222 -d '{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 4d"

#different shape tensor - this one is batch with one element per entry
curl 0.0.0.0:2222 -d '{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: tensor" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 5e"

#text in ndarray - based on moviesentiment example
curl 0.0.0.0:2222 -d '{"data": {"names": ["Text review"],"ndarray": ["this film has bad actors"]}}' -H "Ce-Inferenceservicename: moviesentiment" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 6f"
curl 0.0.0.0:2222 -d '{"data":{"names":["t0","t1"],"ndarray":[[0.5,0.5]]}}' -H "Ce-Inferenceservicename: moviesentiment" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Requestid: 6f"

# escape characters below make the strData one big string
curl 0.0.0.0:2222 -d '{"strData":"{\"columns\":[\"DISPO_CD\",\"ENG_CD\",\"HUE_CD\",\"SALE_OFFER_CD\",\"SHADE_CD\",\"TRGTPRCE_MDLGRP_CD\",\"TRGT_CUST_GROUP_CD\",\"TRG_CATG\",\"VIN\",\"calc_cd\",\"category\",\"color\",\"cond_cd\",\"country\",\"cust_cd\",\"default_cond_cd\",\"dispo_date\",\"dispo_day\",\"drivetype\",\"floor_price\",\"mlge_arriv\",\"mlge_dispo\",\"model\",\"modelyr\",\"region\",\"saleloc\",\"series_cd\",\"sys_enter_date\",\"tag\",\"target_price\",\"v47\",\"v62\",\"v64\",\"vehvalue\",\"warranty_age\",\"wrstdt\",\"wsd\"],\"index\":[0],\"data\":[[3,\"L\",\"RD\",\"CAO\",\"DK\",41,1,\"RTR\",\"MAJ6P1CL3JC166908\",null,\"RPO\",\"RR\",5,\"A\",7,3,\"2018-07-11\",6766,null,0.0,2013,2013,\"ECO\",2018,1,\"C63\",\"P1C\",\"2018-06-16\",null,0.0,\"5\",null,\"5\",\"ecosport\",146.0,\"2018-02-15\",26750.56]]}"}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: strdata" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 7g"

#kfserving tensor - iris (batch of two)
curl 0.0.0.0:2222 -d '{"instances": [[6.8,  2.8,  4.8,  1.4],[6.0,  3.4,  4.5,  1.6]]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: iris-kf" -H "Ce-Endpoint: default" -H "Ce-Requestid: 8h"
curl 0.0.0.0:2222 -d '{"predictions": [1,1]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.response" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: iris-kf" -H "Ce-Endpoint: default" -H "Ce-Requestid: 8h"
curl 0.0.0.0:2222 -d '{"data": {"feature_score": null, "instance_score": null, "is_outlier": [1, 1]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.outlier" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: iris-kf" -H "Ce-Endpoint: default" -H "Ce-id: 8h"

#cifar10 image - kfserving (possibly old)
curl 0.0.0.0:2222 --data-binary "@cifardata.json" -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.request" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 9i" -H 'CE-SpecVersion: 0.2' -v
curl 0.0.0.0:2222 -d '{"predictions":[2]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.response" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 9i"
curl 0.0.0.0:2222 -d '{"data": {"feature_score": null, "instance_score": null, "is_outlier": [1]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.outlier" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 9i"

#dummy tabular example
curl 0.0.0.0:2222 -d '{"data": {"names": ["dummy feature"],"ndarray": [1.0]}}' -H "Ce-Inferenceservicename: dummytabular" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 10j"

#jsonData example
curl 0.0.0.0:2222 -d "{\"jsonData\": {\"input\": \"{'input': '[[53  4  0  2  8  4  2  0  0  0 60  9]]'}\"},\"meta\": {}}" -H "Ce-Inferenceservicename: jsonexample" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 11k"
curl 0.0.0.0:2222 -d "{\"jsonData\": {\"input\": \"{'input': '[[53  4  0  2  8  4  2  0  0  0 60  9]]'}\"},\"meta\": {}}" -H "Ce-Inferenceservicename: jsonexample" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Requestid: 11k"

# tabular input
curl 0.0.0.0:2222 -d  '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: tensor" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 2z1"

# multi-output (only first output handled for now)
curl 0.0.0.0:2222 -d '{"model_name": "simple", "model_version": "1", "outputs": [{"name": "OUTPUT0", "datatype": "INT32", "shape": [1, 16], "data": [2, 4, 6, 8, 10, 12, 14, 16, 8, 20, 22, 24, 26, 28, 30, 32]}, {"name": "OUTPUT1", "datatype": "INT32", "shape": [1, 16], "data": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}]}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: tensor" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Requestid: 2z1"


# image input
curl 0.0.0.0:2222 -d  '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,1,2,8]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: tensor" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 2z2"

# bytes input - doesn't presently work
#curl 0.0.0.0:2222 -d  '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"BYTES","shape":[16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: tensor" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 2z3"


# income classifier
curl 0.0.0.0:2222 -d  '{"data":{"names":["Age","Workclass","Education","Marital Status","Occupation","Relationship","Race","Sex","Capital Gain","Capital Loss","Hours per week","Country"],"ndarray":[[53,4,0,2,8,4,2,0,0,0,60,9]]}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: income" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Endpoint: default" -H "Ce-Requestid: 2z4"
curl 0.0.0.0:2222 -d  '{"data":{"names":["t:0","t:1"],"ndarray":[[0.8538818809164035,0.14611811908359656]]},"meta":{"requestPath":{"income-container":"seldonio\/sklearnserver:1.7.0"}}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: income" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Namespace: seldon" -H "Ce-Endpoint: default" -H "Ce-Requestid: 2z4"

# mix of one_hot, categorical and float
curl 0.0.0.0:2222 -d  '{"data":{"names":["dummy_one_hot_1","dummy_one_hot_2","dummy_categorical","dummy_float"],"ndarray":[[0,1,0,2.54]]}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: dummy" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Endpoint: default" -H "Ce-Requestid: 2z5"
curl 0.0.0.0:2222 -d  '{"data":{"names":["dummy_proba_0","dummy_proba_1","dummy_float"],"ndarray":[[0.8538818809164035,0.14611811908359656,3.65]]}}' -H "Content-Type: application/json" -H "Ce-Inferenceservicename: dummy" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Namespace: seldon" -H "Ce-Endpoint: default" -H "Ce-Requestid: 2z5"

#cifar10 image - seldon
curl 0.0.0.0:2222 --data-binary "@cifardata.json" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 2z6" -v
curl 0.0.0.0:2222 -d '{"predictions":[[1.26448515e-6,4.88145879e-9,1.51533219e-9,8.49055848e-9,5.51306611e-10,1.16171928e-9,5.77288495e-10,2.88396933e-7,0.000614895718,0.999383569]]}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 2z6"
curl 0.0.0.0:2222 -d '{"data": {"feature_score": null, "instance_score": null, "is_outlier": [1]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.outlier" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 2z6"

#iris seldon - batch
curl 0.0.0.0:2222 -d '{"data":{"names":["Sepal length","Sepal width","Petal length","Petal Width"],"ndarray":[[6.8,2.8,4.8,1.4],[6.1,3.4,4.5,1.6]]}}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: iris" -H "Ce-Endpoint: default" -H "Ce-id: 2z7" -v

#iris seldon - not batch
curl 0.0.0.0:2222 -d '{"data":{"names":["Sepal length","Sepal width","Petal length","Petal Width"],"ndarray":[[6.8,2.8,4.8,1.4]]}}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: iris" -H "Ce-Endpoint: default" -H "Ce-id: 2z8" -v

#kfserving income
curl 0.0.0.0:2222 -d '{"instances":[[39, 7, 1, 1, 1, 1, 4, 1, 2174, 0, 40, 9]]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: income-kf" -H "Ce-Endpoint: default" -H "Ce-Requestid: 2z9"

#v2 protocol seldon triton tf cifar10
curl 0.0.0.0:2222 --data-binary "@truck-v2.json" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: tfcifar10" -H "Ce-Endpoint: default" -H "Ce-id: 2z10" -v
curl 0.0.0.0:2222 -d '{"model_name":"cifar10","model_version":"1","outputs":[{"data":[1.2644851494769682e-6,4.881458792738158e-9,1.5153321930583274e-9,8.490558478513321e-9,5.513066114737342e-10,1.1617192763324624e-9,5.772884947852219e-10,2.8839693300142244e-7,0.0006148957181721926,0.9993835687637329],"datatype":"FP32","name":"fc10","shape":[1,10]}]}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: tfcifar10" -H "Ce-Endpoint: default" -H "Ce-Requestid: 2z10"

#v2 protocol seldon iris batch
curl 0.0.0.0:2222 -d '{"model_name":"iris","model_version":"1","inputs":[{"name":"Sepal length","shape":[2,1],"datatype":"FP32","data":[6.8,6.1]},{"name":"Sepal width","shape":[2,1],"datatype":"FP32","data":[2.8,3.4]},{"name":"Petal length","shape":[2,1],"datatype":"FP32","data":[4.8,4.5]},{"name":"Petal width","shape":[2,1],"datatype":"FP32","data":[1.4,1.6]}]}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: iris" -H "Ce-Endpoint: default" -H "Ce-id: 2z11" -v
curl 0.0.0.0:2222 -d '{"model_name":"iris","model_version":"1","outputs":[{"data":[0.1,0.9,0.6,0.4],"datatype":"FP32","name":"dummy_proba_output_name","shape":[2,2]}]}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: iris" -H "Ce-Endpoint: default" -H "Ce-id: 2z11" -v

#v2 protocol seldon triton tf cifar10 batch
curl 0.0.0.0:2222 --data-binary "@batch-truck-v2.json" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: tfcifar10" -H "Ce-Endpoint: default" -H "Ce-id: 2z12" -v
curl 0.0.0.0:2222 -d '{"model_name":"cifar10","model_version":"1","outputs":[{"data":[1.2644841262954287e-6,4.881449466864751e-9,1.5153321930583274e-9,8.490558478513321e-9,5.513055567618608e-10,1.1617192763324624e-9,5.772873845621973e-10,2.8839664878432814e-7,0.0006148945540189743,0.9993835687637329,1.2644851494769682e-6,4.881449466864751e-9,1.5153293064784634e-9,8.490558478513321e-9,5.513055567618608e-10,1.1617170558864132e-9,5.772873845621973e-10,2.8839664878432814e-7,0.0006148945540189743,0.9993835687637329],"datatype":"FP32","name":"fc10","shape":[2,10]}]}' -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Namespace: seldon" -H "Ce-Inferenceservicename: tfcifar10" -H "Ce-Endpoint: default" -H "Ce-Requestid: 2z12"
