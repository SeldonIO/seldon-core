SELDON_CORE_DIR=../..
SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE=alibiexplainer

.PHONY: get_apis
get_apis:
	# Protobuf
	cp ${SELDON_CORE_DIR}/proto/prediction.proto alibiexplainer/proto/
	$(MAKE) -C ${SELDON_CORE_DIR}/proto/tensorflow/ create_protos
	cp -r $(SELDON_CORE_DIR)/proto/tensorflow/tensorflow \
		alibiexplainer/proto/
	$(MAKE) -C ${SELDON_CORE_DIR}/proto/tensorflow clean

.PHONY: build_apis
build_apis: get_apis
	# Protobuf
	cd alibiexplainer && python \
		-m grpc.tools.protoc \
		-I./ \
		-I./proto/ \
		--python_out=./ \
		--grpc_python_out=./ \
		--mypy_out=./ \
		./proto/prediction.proto
	## We build TF's protobufs as well in case the tensorflow package
	## is not found
	cd alibiexplainer/proto && python \
		-m grpc.tools.protoc \
		-I./ \
		-I../ \
		--python_out=./ \
		./tensorflow/core/framework/*.proto
	sed -i "s/from tensorflow/from alibiexplainer.proto.tensorflow/" alibiexplainer/proto/*.py
	sed -i "s/from tensorflow.core.framework/from ./" \
		$(addprefix alibiexplainer/proto/tensorflow/core/framework/, \
			resource_handle_pb2.py tensor_pb2.py tensor_shape_pb2.py types_pb2.py)
	sed -i "s/from proto/from alibiexplainer.proto/g" alibiexplainer/proto/prediction_pb2_grpc.py

dev_install:
	pip install -e .[test]

test: type_check
	pytest -W ignore

type_check:
	mypy --ignore-missing-imports alibiexplainer

docker-build: build_apis
	docker build --file=Dockerfile --force-rm=true -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push seldonio/${IMAGE}:${VERSION}

kind_load: docker-build
	kind load docker-image seldonio/${IMAGE}:${VERSION}

# password can be found at: https://connect.redhat.com/project/3987291/view
redhat-image-scan:
	docker pull seldonio/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_alibi_explain} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag seldonio/${IMAGE}:${VERSION} scan.connect.redhat.com/ospid-02f3e15b-c16f-4353-affa-61d5f3c6408b/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/ospid-02f3e15b-c16f-4353-affa-61d5f3c6408b/${IMAGE}:${VERSION}

clean:
	rm -rf kfserving


#
# Test Tabular Explanations
#

run_predictor_adult:
	docker run --rm -d --name "sklearnserver"  -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"gs://seldon-models/sklearn/income/model-0.23.2"}]' seldonio/sklearnserver_rest:${VERSION}

curl_predict_adult:
	curl -d '{"data": {"ndarray":[[39, 7, 1, 1, 1, 1, 4, 1, 2174, 0, 40, 9]]}}'    -X POST http://localhost:5000/api/v1.0/predictions    -H "Content-Type: application/json"

run_explainer_adult:
	python -m alibiexplainer --model_name adult --protocol seldon.http --storage_uri gs://seldon-models/sklearn/income/alibi/0.4.0 --predictor_host localhost:5000 AnchorTabular

run_explainer_adult_docker:
	docker run --rm -d --name "explainer" --network=host -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name adult --protocol seldon.http --storage_uri gs://seldon-models/sklearn/income/alibi/0.4.0 --predictor_host localhost:5000 AnchorTabular

curl_explain_adult:
	curl -d '{"data": {"ndarray":[[39, 7, 1, 1, 1, 1, 4, 1, 2174, 0, 40, 9]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"

cleanup_adult:
	docker rm -f sklearnserver
	docker rm -f explainer

#
# Test Text Explanations
#

run_predictor_movie:
	docker run --rm -d --name "sklearnserver"  -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"gs://seldon-models/sklearn/moviesentiment"}]' seldonio/sklearnserver_rest:${VERSION}

curl_predict_movie:
	curl -d '{"data": {"ndarray":["a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification"]}}'    -X POST http://localhost:5000/api/v1.0/predictions    -H "Content-Type: application/json"

run_explainer_movie:
	python -m alibiexplainer --model_name adult --protocol seldon.http --predictor_host localhost:5000 AnchorText

run_explainer_movie_docker:
	docker run --rm -d --name "explainer" --network=host -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name adult --protocol seldon.http --predictor_host localhost:5000 AnchorText

curl_explain_movie:
	curl -d '{"data": {"ndarray":["a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification"]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"

cleanup_movie:
	docker rm -f sklearnserver
	docker rm -f explainer

#
# Test Image Explanation
#

models/resnet32:
	mkdir -p models && gsutil cp -r gs://seldon-models/tfserving/cifar10/resnet32 models


run_predictor_image:
	docker run --name tfserver --rm -d -p 8501:8501 -p 8500:8500 -v "${PWD}/models:/models" -e MODEL_NAME=resnet32 tensorflow/serving


curl_predict_image:
	curl -d @./input.json  -X POST http://localhost:8501/v1/models/resnet32:predict    -H "Content-Type: application/json"


run_explainer_image:
	python -m alibiexplainer --model_name resnet32 --protocol tensorflow.http --storage_uri gs://seldon-models/tfserving/imagenet/alibi/0.4.0 --predictor_host localhost:8501 AnchorImages

run_explainer_image_docker:
	docker run --rm -d --name "explainer" --network=host -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name resnet32 --protocol tensorflow.http --storage_uri gs://seldon-models/tfserving/imagenet/alibi/0.4.0 --predictor_host localhost:8501 AnchorImages

curl_explain_image:
	curl -d @./input.json  -X POST http://localhost:8080/v1/models/resnet32:explain    -H "Content-Type: application/json"

cleanup_image:
	docker rm -f tfserver
	docker rm -f explainer


#
# Test Kernel Shap Explanation
#


run_explainer_kernelshap:
	python -m alibiexplainer --model_name adult --protocol seldon.http --storage_uri gs://seldon-models/sklearn/adult_shap/kernelshap/py36 --predictor_host localhost:5000 KernelShap

#
# Docker image is py36 - need to dill.save in py36 the explainer otherwise
# 


run_explainer_kernelshap_docker:
	docker run --rm -d --name "explainer" --network=host -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name adult --protocol seldon.http --storage_uri gs://seldon-models/sklearn/adult_shap/kernelshap/py36 --predictor_host localhost:5000 KernelShap

cleanup_kernelshap:
	docker rm -f sklearnserver
	docker rm -f explainer


#
# Test Integrated Gradients
#

run_explainer_integratedgradients:
	python -m alibiexplainer --model_name adult --protocol seldon.http --storage_uri gs://seldon-models/keras/imdb IntegratedGradients IntegratedGradients


run_explainer_integratedgradients_docker:
	docker run --rm -d --name "explainer" --network=host -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name adult --protocol seldon.http --storage_uri gs://seldon-models/keras/imdb IntegratedGradients

curl_explain_imdb:
	curl -d '{"data": {"ndarray":[[   0,    0,    0,    0,    0,    0,    0,    0,    0,    0,    0, 0,    0,    0,    0,    0,    0,    0,    0,    0,    0,    0, 0,    0,    0,    0,    0,    0,    0,    0,    0,    0,    1, 591,  202,   14,   31,    6,  717,   10,   10,    2,    2,    5, 4,  360,    7,    4,  177, 5760,  394,  354,    4,  123,    9, 1035, 1035, 1035,   10,   10,   13,   92,  124,   89,  488, 7944, 100,   28, 1668,   14,   31,   23,   27, 7479,   29,  220,  468, 8,  124,   14,  286,  170,    8,  157,   46,    5,   27,  239, 16,  179,    2,   38,   32,   25, 7944,  451,  202,   14,    6, 717]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"
