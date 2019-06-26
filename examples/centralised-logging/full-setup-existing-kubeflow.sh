# Assumes existing cluster with kubeflow
# Will put services behind kubeflow istio gateway
# Before running modify ./kubeflow/seldon-analytics-kubeflow.yaml for KF gateway
# KF gateway IP is `kubectl get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`

./kubeflow/knative-setup-existing-istio.sh

sleep 5

kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller

kubectl rollout status -n kube-system deployment/tiller-deploy

helm install --name seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system --set istio.gateway="kubeflow-gateway.kubeflow.svc.cluster.local" --set istio.enabled="true"

kubectl rollout status -n seldon-system statefulset/seldon-operator-controller-manager

sleep 5

helm install --name seldon-single-model ../../helm-charts/seldon-single-model/ --set engine.env.LOG_MESSAGES_EXTERNALLY="true" --set model.annotations."seldon\.io/istio-gateway"="kubeflow-gateway.kubeflow.svc.cluster.local"

kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite
helm install --name seldon-core-loadtesting ../../helm-charts/seldon-core-loadtesting/ --set locust.host=http://seldon-single-model-seldon-single-model:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=0 --set locust.maxWait=0 --set replicaCount=1

helm install --name seldon-core-analytics ../../helm-charts/seldon-core-analytics/ -f ./kubeflow/seldon-analytics-kubeflow.yaml

helm install --name elasticsearch elasticsearch --version 7.1.1 --namespace=logs --set service.type=ClusterIP --set antiAffinity="soft" --repo https://helm.elastic.co
kubectl rollout status statefulset/elasticsearch-master -n logs

helm install fluentd-elasticsearch --name fluentd --namespace=logs -f fluentd-values.yaml --repo https://kiwigrid.github.io
helm install kibana --version 7.1.1 --name=kibana --namespace=logs --set service.type=ClusterIP -f ./kubeflow/kibana-values.yaml --repo https://helm.elastic.co

kubectl apply -f ./kubeflow/virtualservice-kibana.yaml

kubectl rollout status deployment/kibana-kibana -n logs

kubectl apply -f ./request-logging/seldon-request-logger.yaml
kubectl label namespace default knative-eventing-injection=enabled
sleep 3
kubectl -n default get broker default
kubectl apply -f ./request-logging/trigger.yaml

echo 'grafana running at:'
echo "$(kubectl get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')/grafana/"
echo 'kibana running at:'
echo "$(kubectl get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')/kibana/"