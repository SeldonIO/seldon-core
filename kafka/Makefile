DOCKER_IMAGE_NAME=seldonio/kafka-core
DOCKER_IMAGE_VERSION=0.1

build_docker_image:
	mkdir -p .m2 && docker build --force-rm=true -t $(PRIVATE_REPO)$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .
push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)

start_kafka: kafka.json
	kubectl apply -f kafka.json
stop_kafka: kafka.json
	kubectl delete --ignore-not-found=true -f kafka.json


#update_proto:
#	cp -v ../proto/seldonengine.proto src/main/proto/ 
#	cp -v ../proto/prediction.proto src/main/proto/ 

port_forward_kafka:
	POD_NAME=$$(kubectl --namespace default get pod -l app=kafka -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl port-forward $${POD_NAME} 9093:9093

