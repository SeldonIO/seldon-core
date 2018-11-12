
.PHONY: test
test:
	PYTHONPATH=../../notebooks pytest -s

.PHONY: port-forward
port-forward:
	kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-aserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8002:8080
