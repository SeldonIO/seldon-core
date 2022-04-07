QUERY='query=mygauge{deployment_name=~"echo",namespace=~"seldon"}'
QUERY_URL=http://seldon-monitoring-prometheus.seldon-system.svc.cluster.local:9090/api/v1/query

kubectl run --quiet=true -it --rm curlmetrics --image=radial/busyboxplus:curl --restart=Never -- \
    curl --data-urlencode ${QUERY} ${QUERY_URL}
