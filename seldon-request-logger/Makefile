build_image:
	docker build . -t seldonio/seldon-request-logger:latest

push_image:
	docker push seldonio/seldon-request-logger:latest

build_image_cifar:
	docker build -f cifar10.Dockerfile . -t seldonio/seldon-request-logger-cifar:latest

push_image_cifar:
	docker push seldonio/seldon-request-logger-cifar:latest

run_container:
	docker run -p 2222:8080 seldonio/seldon-request-logger:latest

run_local: export FLASK_RUN_PORT = 2222
run_local: export FLASK_APP = app.py
run_local:
	cd app && flask run

run_local_cifar: export FLASK_RUN_PORT = 2222
run_local_cifar: export FLASK_APP = app.py
run_local_cifar:
	cd cifar && flask run

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
