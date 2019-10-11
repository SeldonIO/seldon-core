# Installs full centralised logging and request logging setup to fresh minikube
# Recommended minikube params --memory 10240 --cpus 6 --disk-size=30g

./request-logging/knative-setup.sh

sleep 5

kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller

kubectl rollout status -n kube-system deployment/tiller-deploy

helm install --name seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system --set engine.logMessagesExternally="true"

kubectl rollout status -n seldon-system deployment/seldon-controller-manager

sleep 5

helm install --name seldon-single-model ../../helm-charts/seldon-single-model/ --set engine.env.LOG_MESSAGES_EXTERNALLY="true"

kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite
helm install --name seldon-core-loadtesting ../../helm-charts/seldon-core-loadtesting/ --set locust.host=http://seldon-single-model-seldon-single-model-seldon-single-model:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=0 --set locust.maxWait=0 --set replicaCount=1



helm install --name elasticsearch elasticsearch --version 7.1.1 --namespace=logs -f elastic-minikube.yaml --repo https://helm.elastic.co
kubectl rollout status statefulset/elasticsearch-master -n logs
kubectl patch svc elasticsearch-master -n logs -p '{"spec": {"type": "LoadBalancer"}}'

helm install fluentd-elasticsearch --name fluentd --namespace=logs -f fluentd-values.yaml --repo https://kiwigrid.github.io
helm install kibana --version 7.1.1 --name=kibana --namespace=logs --set service.type=NodePort --repo https://helm.elastic.co

kubectl rollout status deployment/kibana-kibana -n logs

kubectl apply -f ./request-logging/seldon-request-logger.yaml
kubectl label namespace default knative-eventing-injection=enabled
sleep 3
kubectl -n default get broker default
kubectl apply -f ./request-logging/trigger.yaml

echo 'kibana running at:'
echo $(minikube ip)":"$(kubectl get svc kibana-kibana -n logs -o=jsonpath='{.spec.ports[?(@.port==5601)].nodePort}')