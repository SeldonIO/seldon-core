redis-memonly.json: redis-memonly.json.in
	@cat redis-memonly.json.in > redis-memonly.json && echo "created $@"

start_redis: redis-memonly.json
	kubectl create -f redis-memonly.json
stop_redis: redis-memonly.json
	kubectl delete -f redis-memonly.json

redis_cli:
	kubectl exec -it redis -- bash -c 'redis-cli'

redis_port_forward:
	kubectl port-forward redis 6379

clean:
	@rm -fv redis-memonly.json

