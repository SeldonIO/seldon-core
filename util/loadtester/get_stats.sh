if [ "$#" -ne 2 ]; then
    echo "Illegal number of parameters: <experiment> <rest|grpc>"
fi

EXPERIMENT=$1
TYPE=$2

MASTER=`kubectl get pod -l name=locust-master-1 -o jsonpath='{.items[0].metadata.name}'`

kubectl cp ${MASTER}:stats_distribution.csv ${EXPERIMENT}_${TYPE}_stats_distribution.csv
kubectl cp ${MASTER}:stats_requests.csv ${EXPERIMENT}_${TYPE}_stats_requests.csv

