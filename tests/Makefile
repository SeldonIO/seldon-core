DEPLOYMENT_JSON=deployment_mean_classifier.json
DEPLOYMENT_ID=`cat config/${DEPLOYMENT_JSON} | jq -r .id`

#
# Manage REST Endpoints
#

CLUSTER_MANAGER_ENDPOINT:=$(shell cat CLUSTER_MANAGER_ENDPOINT)
CLUSTER_MANAGER_CLIENT_SECRET=$(shell cat ../cluster-manager/cluster-manager-client-secret.txt)
API_ENDPOINT:=$(shell cat API_ENDPOINT)
API_CLIENT_KEY=$(shell cat config/$(DEPLOYMENT_JSON) | jq -r .oauth_key)
API_CLIENT_SECRET=$(shell cat config/$(DEPLOYMENT_JSON) | jq -r .oauth_secret)

cm_update_endpoint_for_localhost:
	@echo "localhost:8081" > CLUSTER_MANAGER_ENDPOINT && cat CLUSTER_MANAGER_ENDPOINT
cm_update_endpoint_for_minikube:
	@MINIKUBE_IP=$$(minikube ip) && echo "$${MINIKUBE_IP}:30030" > CLUSTER_MANAGER_ENDPOINT && cat CLUSTER_MANAGER_ENDPOINT
cm_update_endpoint_for_loadbalancer:
	@API_IP=$$(kubectl get svc seldon-cluster-manager  -o jsonpath="{.status.loadBalancer.ingress[0].ip}") && echo "$${API_IP}:8080" > CLUSTER_MANAGER_ENDPOINT && cat CLUSTER_MANAGER_ENDPOINT

api_update_endpoint_for_localhost:
	@echo "localhost:8080" > API_ENDPOINT && cat API_ENDPOINT
api_update_endpoint_for_minikube:
	@MINIKUBE_IP=$$(minikube ip) && echo "$${MINIKUBE_IP}:30032" > API_ENDPOINT && cat API_ENDPOINT
api_update_endpoint_for_loadbalancer:
	@API_IP=$$(kubectl get svc seldon-apiserver  -o jsonpath="{.status.loadBalancer.ingress[0].ip}") && echo "$${API_IP}:8080" > API_ENDPOINT && cat API_ENDPOINT


#
# API Resource URIs
#

EP_PING=/ping
EP_VERSION=/version
EP_TOKEN=/oauth/token
EP_AUTHPING=/api/v1/authping
EP_NAMESPACES=/api/v1/namespaces
EP_DEPLOYMENTS=/api/v1/deployments
EP_DOCKER_REGISTRY_SECRETS=/api/v1/docker-registry-secrets
EP_FEEDBACK=/api/v0.1/feedback
EP_PREDICTION=/api/v0.1/predictions



CM_ACCESS_TOKEN=$(shell curl -s -H "Accept: application/json" client:$(CLUSTER_MANAGER_CLIENT_SECRET)@$(CLUSTER_MANAGER_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials|jq -r '.access_token')

API_ACCESS_TOKEN=$(shell curl -s -H "Accept: application/json" ${API_CLIENT_KEY}:$(API_CLIENT_SECRET)@$(API_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials|jq -r '.access_token')

notarget:
	@echo need target



#
# Cluster Manager API calls
#

cm_check_container_logs:
	kubectl logs `kubectl get pods | grep cluster-manager | cut -f1 -d' '` -c seldon-cluster-manager-container

cm_version:
	curl -s -i $(CLUSTER_MANAGER_ENDPOINT)$(EP_VERSION)
cm_ping:
	curl -s -i $(CLUSTER_MANAGER_ENDPOINT)$(EP_PING)

cm_get_namespaces:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(CM_ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Accept: application/json" -H "Authorization: Bearer $(CM_ACCESS_TOKEN)" $(CLUSTER_MANAGER_ENDPOINT)$(EP_NAMESPACES)
cm_authping:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(CM_ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Accept: application/json" -H "Authorization: Bearer $(CM_ACCESS_TOKEN)" http://$(CLUSTER_MANAGER_ENDPOINT)$(EP_AUTHPING)
cm_get_token:
	curl -i  -H "Accept: application/json" client:$(CLUSTER_MANAGER_CLIENT_SECRET)@$(CLUSTER_MANAGER_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials

cm_create_deployment:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(CM_ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H 'Content-Type: application/json' -H "Authorization: Bearer $(CM_ACCESS_TOKEN)" -X "POST" -d @config/$(DEPLOYMENT_JSON)  $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS) | tee out.log

cm_get_deployment:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(CM_ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Authorization: Bearer $(CM_ACCESS_TOKEN)" -X "GET" $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS)/1 | tee out.log

cm_check_deployment_ready:
		curl -s -H "Authorization: Bearer $(CM_ACCESS_TOKEN)" -X "GET" $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS)/1 | python scripts/check_deployment_ready.py

cm_delete_deployment:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(CM_ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Authorization: Bearer $(CM_ACCESS_TOKEN)" -X "DELETE" $(CLUSTER_MANAGER_ENDPOINT)$(EP_DEPLOYMENTS)/${DEPLOYMENT_ID} | tee out.log


#
# API Calls
#

api_ping:
	curl -s -i $(API_ENDPOINT)$(EP_PING)


api_get_token:
	curl -v -H "Accept: application/json" ${API_CLIENT_KEY}:$(API_CLIENT_SECRET)@$(API_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials


api_post_test:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(API_ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -v -i -X POST -H "Content-Type:application/json" -H "Accept: application/json" -H "Authorization: Bearer $(API_ACCESS_TOKEN)" $(API_ENDPOINT)$(EP_PREDICTION) -d '{"request":{"features":["a"],"ndarray":[[1.0],[2.0]]}}'

api_feedback_test:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(API_ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -v -i -X POST -H "Content-Type:application/json" -H "Accept: application/json" -H "Authorization: Bearer $(API_ACCESS_TOKEN)" $(API_ENDPOINT)$(EP_FEEDBACK) -d '{"request":{"features":["a"],"ndarray":[[1.0],[2.0]]},"response":{"meta":{"puid":"vmjfejn90adjniol7b8ghkk8un","routing":{"0": 0}},"response": {"features": ["proba"],"tensor": {"shape":[1,2],"values":[0.12823373759251927,0.2856375705089441]}}},"reward":1}'


