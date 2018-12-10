# Load testing

# Local usage

locust -f scripts/mnist_grpc_locust.py --master --host=0.0.0.0:5000 --no-web

OAUTH_ENABLED=false API_ENDPOINT=internal locust -f scripts/mnist_grpc_locust.py --slave --host=0.0.0.0:5000

docker run --rm -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"name":"cols","type":"FLOAT","value":"10"},{"name":"rows","type":"FLOAT","value":1}]' seldonio/shaped_random_model_grpc:0.1


