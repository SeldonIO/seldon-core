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
curl 0.0.0.0:2222 -d '{"instances": [[6.8,  2.8,  4.8,  1.4],[6.0,  3.4,  4.5,  1.6]]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.request" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: iris" -H "Ce-Requestid: 8h"
curl 0.0.0.0:2222 -d '{"predictions": [1,1]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.response" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: iris" -H "Ce-Requestid: 8h"
curl 0.0.0.0:2222 -d '{"data": {"feature_score": null, "instance_score": null, "is_outlier": [1, 1]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.outlier" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: iris" -H "Ce-id: 8h"

#cifar10 image
curl 0.0.0.0:2222 --data-binary "@cifardata.json" -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.request" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 9i" -H 'CE-SpecVersion: 0.2' -v
curl 0.0.0.0:2222 -d '{"predictions":[2]}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.response" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 9i"
curl 0.0.0.0:2222 -d '{"data": {"feature_score": null, "instance_score": null, "is_outlier": [1]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}' -H "Content-Type: application/json" -H "Ce-Type: org.kubeflow.serving.inference.outlier" -H "Ce-Namespace: default" -H "Ce-Inferenceservicename: cifar10" -H "Ce-Endpoint: default" -H "Ce-id: 9i"

#dummy tabular example
curl 0.0.0.0:2222 -d '{"data": {"names": ["dummy feature"],"ndarray": [1.0]}}' -H "Ce-Inferenceservicename: dummytabular" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 10j"

#jsonData example
curl 0.0.0.0:2222 -d "{\"jsonData\": {\"input\": \"{'input': '[[53  4  0  2  8  4  2  0  0  0 60  9]]'}\"},\"meta\": {}}" -H "Ce-Inferenceservicename: jsonexample" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.request" -H "Ce-Requestid: 11k"
curl 0.0.0.0:2222 -d "{\"jsonData\": {\"input\": \"{'input': '[[53  4  0  2  8  4  2  0  0  0 60  9]]'}\"},\"meta\": {}}" -H "Ce-Inferenceservicename: jsonexample" -H "Content-Type: application/json" -H "Ce-Type: io.seldon.serving.inference.response" -H "Ce-Requestid: 11k"
