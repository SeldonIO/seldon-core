VERSION=1.3
IMAGE_BASE=seldonio/mock_classifier

build_rest:
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python36:0.12-SNAPSHOT ${IMAGE_BASE}_rest:${VERSION}

push_rest:
	docker push ${IMAGE_BASE}_rest:${VERSION}

build_grpc:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python36:0.11-SNAPSHOT ${IMAGE_BASE}_grpc:${VERSION}

push_grpc:
	docker push ${IMAGE_BASE}_grpc:${VERSION}

run_rest_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && seldon-core-microservice --service-type MODEL MeanClassifier REST
