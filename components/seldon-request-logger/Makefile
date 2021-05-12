SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
REPO := seldonio

.EXPORT_ALL_VARIABLES:
CLIENT_ID=sd-api
OIDC_USERNAME=admin@seldon.io
OIDC_PASSWORD=xxxxxx
OIDC_SCOPES=openid profile email groups
DEPLOY_API_HOST=http://xx.xx.xx.xx/seldon-deploy/api/v1alpha1
OIDC_PROVIDER=http://xx.xx.xx.xx/auth/realms/deploy-realm
ELASTICSEARCH_HOST=localhost
ELASTICSEARCH_PORT=9200

#to test metadata, set xxs above then run deploy_test_model and create_dummy_metadata
#then run_elastic in another window
#then run_local
#then curl_metadata

deploy_test_model:
	./testing/deploy_model.sh

create_dummy_metadata:
	python testing/create_dummy_metadata.py

run_container:
	docker run -p 2222:8080 seldonio/seldon-request-logger:latest

run_elastic:
	docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch-oss:7.6.0

run_local: export FLASK_RUN_PORT = 2222
run_local: export FLASK_APP = default_logger.py
run_local:
	cd app && flask run

curl_metadata:
	curl -X GET -k -v http://localhost:2222/metadata -H "Content-Type: application/json"

#see test.sh for more test data
test:
	curl -v "http://localhost:2222/" \
      -X POST \
      -H "X-B3-Flags: 1" \
      -H 'CE-SpecVersion: 0.2' \
      -H "CE-Type: io.seldon.serving.inference.request" \
      -H "CE-Time: 2018-04-05T03:56:24Z" \
      -H "CE-ID: 45a8b444-3213-4758-be3f-540bf93f85ff" \
      -H "CE-Source: dev.knative.example" \
      -H 'Content-Type: application/json' \
      -d '{"request": {"meta": {"puid": "71dlk7k1rhmci0cd8g5rmeolmn", "tags": {}, "routing": {}, "requestPath": {}, "metrics": []}, "data": {"names": ["f0", "f1"], "ndarray": [[0.77, 0.63]]}, "date": "2019-06-17T10:59:55.693Z[GMT]"}, "response": {"meta": {"puid": "71dlk7k1rhmci0cd8g5rmeolmn", "tags": {}, "routing": {}, "requestPath": {"classifier": "seldonio/mock_classifier:1.0"}, "metrics": []}, "data": {"names": ["proba"], "ndarray": [[0.09826376903346358]]}, "date": "2019-06-17T10:59:55.696Z[GMT]"}, "sdepName": "seldon-single-model"}'

test_single_dim:
	curl -v "http://localhost:2222/" \
      -X POST \
      -H "X-B3-Flags: 1" \
      -H 'CE-SpecVersion: 0.2' \
      -H "CE-Type: io.seldon.serving.inference.request" \
      -H "CE-Time: 2018-04-05T03:56:24Z" \
      -H "CE-ID: 45a8b444-3213-4758-be3f-540bf93f85ff" \
      -H "CE-Source: dev.knative.example" \
      -H 'Content-Type: application/json' \
      -d '{"request": {"meta": {"puid": "71dlk7k1rhmci0cd8g5rmeolmn", "tags": {}, "routing": {}, "requestPath": {}, "metrics": []}, "data": {"names": ["f0", "f1"], "ndarray": [0.77, 0.63]}, "date": "2019-06-17T10:59:55.693Z[GMT]"}, "response": {"meta": {"puid": "71dlk7k1rhmci0cd8g5rmeolmn", "tags": {}, "routing": {}, "requestPath": {"classifier": "seldonio/mock_classifier:1.0"}, "metrics": []}, "data": {"names": ["proba"], "ndarray": [0.09826376903346358]}, "date": "2019-06-17T10:59:55.696Z[GMT]"}, "sdepName": "seldon-single-model"}'


IMAGE=seldon-request-logger

build_image:
	docker build . -t ${REPO}/${IMAGE}:${VERSION}

push_image:
	docker push ${REPO}/${IMAGE}:${VERSION}

kind_load: build_image
	kind load docker-image ${REPO}/${IMAGE}:${VERSION}


#
# RedHat
#

# password can be found at: https://connect.redhat.com/project/3993051/view
redhat-image-scan:
	docker pull ${REPO}/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_request_logger} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${REPO}/${IMAGE}:${VERSION} scan.connect.redhat.com/ospid-62dca88a-015b-4d08-8c52-b709e55a7790/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/ospid-62dca88a-015b-4d08-8c52-b709e55a7790/${IMAGE}:${VERSION}
