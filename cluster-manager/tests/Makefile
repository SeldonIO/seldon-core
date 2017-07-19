CLUSTER_MANAGER_ENDPOINT:=$(shell cat CLUSTER_MANAGER_ENDPOINT)
CLUSTER_MANAGER_CLIENT_SECRET=$(shell cat ../cluster-manager-client-secret.txt)

EP_PING=/ping
EP_VERSION=/version
EP_TOKEN=/oauth/token
EP_AUTHPING=/api/v1/authping
EP_NAMESPACES=/api/v1/namespaces
EP_DEPLOYMENTS=/api/v1/deployments
EP_DOCKER_REGISTRY_SECRETS=/api/v1/docker-registry-secrets

ACCESS_TOKEN=$(shell curl -s -H "Accept: application/json" client:$(CLUSTER_MANAGER_CLIENT_SECRET)@$(CLUSTER_MANAGER_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials|jq -r '.access_token')

notarget:
	@echo need target

update_cluster_manager_endpoint_for_localhost:
	@echo "localhost:8080" > CLUSTER_MANAGER_ENDPOINT && cat CLUSTER_MANAGER_ENDPOINT
update_cluster_manager_endpoint_for_minikube:
	@MINIKUBE_IP=$$(minikube ip) && echo "$${MINIKUBE_IP}:30030" > CLUSTER_MANAGER_ENDPOINT && cat CLUSTER_MANAGER_ENDPOINT

check_cluster_manager_container_logs:
	kubectl logs seldon-cluster-manager -c seldon-cluster-manager-container

do_version:
	curl -s -i $(CLUSTER_MANAGER_ENDPOINT)$(EP_VERSION)
do_ping:
	curl -s -i $(CLUSTER_MANAGER_ENDPOINT)$(EP_PING)


do_get_namespaces:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Accept: application/json" -H "Authorization: Bearer $(ACCESS_TOKEN)" $(CLUSTER_MANAGER_ENDPOINT)$(EP_NAMESPACES)
do_authping:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Accept: application/json" -H "Authorization: Bearer $(ACCESS_TOKEN)" http://$(CLUSTER_MANAGER_ENDPOINT)$(EP_AUTHPING)
do_get_token:
	curl -i  -H "Accept: application/json" client:$(CLUSTER_MANAGER_CLIENT_SECRET)@$(CLUSTER_MANAGER_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials




do_create_deployment:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H 'Content-Type: application/json' -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "POST" -d @deployments/mean.json  $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS) | tee out.log


do_get_deployment:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "GET" $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS)/1 | tee out.log

do_delete_deployment:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "DELETE" $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS)/1 | tee out.log

do_update_deployment:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H 'Content-Type: application/json' -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "PATCH" -d @deployment_4_and_8.json $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS) | tee out.log
do_create_deployment_with_resource_requests.json:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H 'Content-Type: application/json' -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "POST" -d @deployment_4_and_6_with_resource_requests.json $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS) | tee out.log
do_create_deployment_with_canary.json:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H 'Content-Type: application/json' -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "POST" -d @deployment_4_and_6_with_canary.json $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS) | tee out.log



forward_cluster_manager_port:
	@echo "localhost:8080" > CLUSTER_MANAGER_ENDPOINT && cat CLUSTER_MANAGER_ENDPOINT && \
			kubectl port-forward seldon-cluster-manager 8080:8080
forward_zookeeper_port:
	kubectl port-forward zookeeper-1 2181:2181
forward_kubeapi_port:
	kubectl port-forward seldon-cluster-manager 8001:8001


show_all:
	@watch kubectl get all -o wide --show-labels
show_seldon_replicas:
	@watch -n 1 ./show-replicas
show_seldon_pods:
	@watch -n 1 kubectl get pods -l seldon-app=sd-1-c -l seldon-app=sd-1-p
show_seldon_services:
	@watch -n 1 kubectl get services -l seldon-deployment-id=1
show_seldon_deployments:
	@watch -n 1 kubectl get deployments -l seldon-deployment-id=1


#Context
get_contexts:
	kubectl config get-contexts
get_current_context:
	kubectl config current-context
set_context_minikube_namespace_dev1:
	@kubectl config set-context minikube --namespace=dev1
	kubectl config use-context minikube
	@kubectl config get-contexts minikube
set_context_minikube_namespace_default:
	@kubectl config set-context minikube --namespace=default
	@kubectl config use-context minikube
	@kubectl config get-contexts minikube
REMOTE_CLUSTER=cluster-testclient-context
set_context_remote_namespace_default:
	@kubectl config set-context "${REMOTE_CLUSTER}" --namespace=default
	@kubectl config use-context "${REMOTE_CLUSTER}"
	@kubectl config get-contexts "${REMOTE_CLUSTER}"



#weavescope
start_weavescope:
	kubectl create -f scope.json
stop_weavescope:
	kubectl delete -f scope.json
port_forward_weavescope:
	kubectl port-forward $$(kubectl get pod --selector=weave-scope-component=app -o jsonpath='{.items..metadata.name}') 4040

#auth for kubectl to use remote cluster (kubectl is configured to use Application Default Credentials to authenticate to the cluster)
kubectl_gcloud_auth:
	gcloud auth application-default login



do_create_bad_deployment:
	curl -s -i -H 'Content-Type: application/json' -X "POST" -d @bad_deployment.json $(CLUSTER_MANAGER_ENDPOINT)/deployments | tee out.log

check_cluster_manager_client_secret:
	kubectl exec -it seldon-cluster-manager -c seldon-cluster-manager-container -- bash -c 'echo [$${SELDON_CLUSTER_MANAGER_CLIENT_SECRET}]'



#Create docker registry secrets
do_create_seldon_registry_secret:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -i -H 'Content-Type: application/json' -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "POST" -d '{"dockerRegistryDetails":{"password":"r6v6sikq","url":"https://reg.rummblelabs.com:8080","username":"reguser"},"name":"seldon-registry-secret"}' localhost:8080$(EP_DOCKER_REGISTRY_SECRETS)
do_delete_seldon_registry_secret:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -i -H 'Content-Type: application/json' -H "Authorization: Bearer $(ACCESS_TOKEN)" -X "DELETE" localhost:8080$(EP_DOCKER_REGISTRY_SECRETS)/seldon-registry-secret



check_pu_pod:
	PU_POD=$$(kubectl get pods -l seldon-app=sd-1-p -l seldon-track=stable | sed 1d | cut -d' ' -f1 |sed -e 's/^[ \t]*//') && kubectl exec -it $${PU_POD} -c seldon-container-engine -- printenv|grep ENGINE_PREDICTOR|sed -e 's/ENGINE_PREDICTOR=//'|base64 -D

