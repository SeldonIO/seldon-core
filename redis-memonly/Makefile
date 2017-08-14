redis-memonly.json: redis-memonly.json.in
	@cat redis-memonly.json.in > redis-memonly.json && echo "created $@"

start_redis: redis-memonly.json
	kubectl apply -f redis-memonly.json
stop_redis: redis-memonly.json
	kubectl delete --ignore-not-found=true -f redis-memonly.json

redis_cli:
	kubectl exec -it redis -- bash -c 'redis-cli'

port_forward_redis:
	kubectl port-forward redis 6379

clean:
	@rm -fv redis-memonly.json

