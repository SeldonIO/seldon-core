
generated:
	rm -rf generated
	mkdir generated
	kustomize build ../config/lite -o generated

deploy/operator.yaml: generated
	python hack/create_resources.py --input generated --output deploy

clean_1.0.2:
	rm -f deploy/olm-catalog/seldon-operator/1.0.2/seldon-operator.v1.0.2.clusterserviceversion.yaml

deploy/olm-catalog/seldon-operator/1.0.2/seldon-operator.v1.0.2.clusterserviceversion.yaml: 
	operator-sdk generate csv --csv-version 1.0.2 --from-version 0.1.5

.PHONY: clean_1.1.0
clean_1.1.0:
	rm -f deploy/olm-catalog/seldon-operator/1.1.0/seldon-operator.v1.1.0.clusterserviceversion.yaml

deploy/olm-catalog/seldon-operator/1.1.0/seldon-operator.v1.1.0.clusterserviceversion.yaml: 
	operator-sdk generate csv --csv-version 1.1.0 --from-version 1.0.2 --update-crds

scorecard:
	operator-sdk scorecard -o text --bundle deploy/olm-catalog/seldon-operator --kubeconfig ~/.kube/config
