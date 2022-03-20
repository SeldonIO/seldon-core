

.PHONY: deploy-local
deploy-local:
	make -C scheduler start-all

.PHONY: undeploy-local
undeploy-local:
	make -C scheduler stop-all

.PHONY: deploy-k8s
deploy-k8s:
	kubectl create -f k8s/seldon-v2-crds.yaml
	kubectl wait --for condition=established --timeout=60s crd/servers.mlops.seldon.io
	kubectl wait --for condition=established --timeout=60s crd/serverconfigs.mlops.seldon.io
	kubectl create -f k8s/seldon-v2-components.yaml
	kubectl create -f k8s/seldon-v2-servers.yaml

.PHONY: undeploy-k8s
undeploy-k8s:
	kubectl delete -f k8s/seldon-v2-servers.yaml
	kubectl delete -f k8s/seldon-v2-components.yaml
	kubectl delete -f k8s/seldon-v2-crds.yaml
