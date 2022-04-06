QUERY='query=seldon_api_executor_client_requests_seconds_count{deployment_name=~"echo",namespace=~"seldon",method=~"unary"}'
QUERY_URL=http://seldon-monitoring-prometheus.seldon-system.svc.cluster.local:9090/api/v1/query

kubectl run --quiet=true -it --rm curlmetrics --image=radial/busyboxplus:curl --restart=Never -- \
    curl --data-urlencode ${QUERY} ${QUERY_URL}
