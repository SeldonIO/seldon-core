SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE=alibiexplainer

dev_install:
	pip install -e .[test]

test: type_check
	pytest -W ignore

type_check:
	mypy --ignore-missing-imports alibiexplainer

docker-build:
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
	docker run --rm -d --name "sklearnserver"  -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"gs://seldon-models/sklearn/income/model"}]' seldonio/sklearnserver_rest:1.3.0-dev

curl_predict_adult:
	curl -d '{"data": {"ndarray":[[39, 7, 1, 1, 1, 1, 4, 1, 2174, 0, 40, 9]]}}'    -X POST http://localhost:5000/api/v1.0/predictions    -H "Content-Type: application/json"

run_explainer_adult:
	python -m alibiexplainer --model_name adult --protocol seldon.http --storage_uri gs://seldon-models/sklearn/income/alibi/0.4.0 --predictor_host localhost:5000 AnchorTabular

curl_explain_adult:
	curl -d '{"data": {"ndarray":[[39, 7, 1, 1, 1, 1, 4, 1, 2174, 0, 40, 9]]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"


#
# Test Text Explanations
#

run_predictor_movie:
	docker run --rm -d --name "sklearnserver"  -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"type":"STRING","name":"model_uri","value":"gs://seldon-models/sklearn/moviesentiment"}]' seldonio/sklearnserver_rest:1.3.0-dev

curl_predict_movie:
	curl -d '{"data": {"ndarray":["a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification"]}}'    -X POST http://localhost:5000/api/v1.0/predictions    -H "Content-Type: application/json"

run_explainer_movie:
	python -m alibiexplainer --model_name adult --protocol seldon.http --predictor_host localhost:5000 AnchorText

curl_explain_movie:
	curl -d '{"data": {"ndarray":["a visually exquisite but narratively opaque and emotionally vapid experience of style and mystification"]}}'    -X POST http://localhost:8080/api/v1.0/explain    -H "Content-Type: application/json"


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


curl_explain_image:
	curl -d @./input.json  -X POST http://localhost:8080/v1/models/resnet32:explain    -H "Content-Type: application/json"
