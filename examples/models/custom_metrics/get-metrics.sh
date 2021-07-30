
IP=$(kubectl get pods -l seldon-deployment-id=seldon-model -n seldon -o jsonpath='{.items[0].status.podIP}')
kubectl run --quiet=true -it --rm curlmetrics --image=radial/busyboxplus:curl --restart=Never -- \
    curl -s ${IP}:6000/prometheus | grep mycounter_total
