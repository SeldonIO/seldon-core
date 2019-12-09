curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}},"response":{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"ndarray":[[1,2],[3,4]]}},"response":{"data":{"names":["c"],"ndarray":[[7],[8]]}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a"],"ndarray":["test1","test2"]}},"response":{"data":{"names":["c"],"ndarray":[[7],[8]]}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"response":{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"tensor":{"shape":[2,2,1],"values":[1,2,3,4]}}},"response":{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}}' -H "Content-Type: application/json"

# escape characters below make the strData one big string
curl 0.0.0.0:8080 -d '{"request":{"strData":"{\"columns\":[\"DISPO_CD\",\"ENG_CD\",\"HUE_CD\",\"SALE_OFFER_CD\",\"SHADE_CD\",\"TRGTPRCE_MDLGRP_CD\",\"TRGT_CUST_GROUP_CD\",\"TRG_CATG\",\"VIN\",\"calc_cd\",\"category\",\"color\",\"cond_cd\",\"country\",\"cust_cd\",\"default_cond_cd\",\"dispo_date\",\"dispo_day\",\"drivetype\",\"floor_price\",\"mlge_arriv\",\"mlge_dispo\",\"model\",\"modelyr\",\"region\",\"saleloc\",\"series_cd\",\"sys_enter_date\",\"tag\",\"target_price\",\"v47\",\"v62\",\"v64\",\"vehvalue\",\"warranty_age\",\"wrstdt\",\"wsd\"],\"index\":[0],\"data\":[[3,\"L\",\"RD\",\"CAO\",\"DK\",41,1,\"RTR\",\"MAJ6P1CL3JC166908\",null,\"RPO\",\"RR\",5,\"A\",7,3,\"2018-07-11\",6766,null,0.0,2013,2013,\"ECO\",2018,1,\"C63\",\"P1C\",\"2018-06-16\",null,0.0,\"5\",null,\"5\",\"ecosport\",146.0,\"2018-02-15\",26750.56]]}"}}' -H "Content-Type: application/json"





