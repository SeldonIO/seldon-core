DOCKER_IMAGE_NAME=seldonio/zookeeper-k8s
SELDON_ZOOKEEPER_K8S_IMAGE_VERSION=1.0
DOCKER_IMAGE_VERSION=$(SELDON_ZOOKEEPER_K8S_IMAGE_VERSION)

build_docker_image:
	docker build --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .

push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)




zookeeper.json: zookeeper.json.in
	@cat zookeeper.json.in | sed \
		-e  "s|%SELDON_ZOOKEEPER_K8S_IMAGE_VERSION%|$(SELDON_ZOOKEEPER_K8S_IMAGE_VERSION)|" \
		> zookeeper.json && echo "created $@"

start_zookeeper: zookeeper.json
	kubectl create -f zookeeper.json
stop_zookeeper: zookeeper.json
	kubectl delete --ignore-not-found=true -f zookeeper.json

zk_shell:
	kubectl exec -it zookeeper-1 -- /opt/zookeeper/bin/zkCli.sh

port_forward_zk:
	kubectl port-forward zookeeper-1 2181:2181

check:
	@echo srvr | nc localhost  2181

clean:
	@rm -fv zookeeper.json
