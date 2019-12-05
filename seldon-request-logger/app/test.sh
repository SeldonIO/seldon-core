curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}},"response":{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"ndarray":[[1,2],[3,4]]}},"response":{"data":{"names":["c"],"ndarray":[[7],[8]]}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a"],"ndarray":["test1","test2"]}},"response":{"data":{"names":["c"],"ndarray":[[7],[8]]}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"tensor":{"shape":[2,2],"values":[1,2,3,4]}}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"response":{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}}' -H "Content-Type: application/json"

curl 0.0.0.0:8080 -d '{"request":{"data":{"names":["a","b"],"tensor":{"shape":[2,2,1],"values":[1,2,3,4]}}},"response":{"data":{"names":["c"],"tensor":{"shape":[2,1],"values":[5,6]}}}}' -H "Content-Type: application/json"







