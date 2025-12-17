QUERY='query=mygauge{model_name="classifier",namespace="seldon"}'
QUERY_URL=http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090/api/v1/query

kubectl run -n seldon --quiet=true -it --rm curlmetrics-$(date +%s) --image=curlimages/curl:8.6.0 --restart=Never -- \
    curl --data-urlencode ${QUERY} ${QUERY_URL}
