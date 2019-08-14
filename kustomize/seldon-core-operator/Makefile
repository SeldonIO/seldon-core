

base:
	helm convert ../../helm-charts/seldon-core-operator --skip-transformers=image,secret,namePrefix,labels,configMap --destination base --namespace seldon-system
	# Convert seems to lose subresources so we add the crd explicitly
	cp ../../helm-charts/seldon-core-operator/templates/crd.yaml base/seldondeployments.machinelearning.seldon.io-crd.yaml

overlays/reporting:
	helm convert ../../helm-charts/seldon-core-operator --skip-transformers=image,secret,namePrefix,labels,configMap --destination overlays/reporting --namespace seldon-system --set usageMetrics.enabled=true
	rm overlays/reporting/seldon-operator-*
	rm overlays/reporting/seldon-core-operator-sa.yaml
	rm overlays/reporting/seldondeployments.machinelearning.seldon.io-crd.yaml
	rm overlays/reporting/webhook-server-service-svc.yaml
	rm overlays/reporting/seldon-config-cm.yaml
	cp reporting_kustomize.yaml overlays/reporting/kustomization.yaml

all: base overlays/reporting

.PHONY: clean
clean:
	rm -rf base
	rm -rf overlays/reporting

.PHONY: deploy-base
deploy-base:
	kubectl create namespace seldon-system || echo "seldon-system namespace exists" 
	kustomize build base | kubectl apply -f -

.PHONY: undeploy-base
undeploy-base:
	kustomize build base | kubectl delete -f -

.PHONY: deploy-base-with-reporting
deploy-base-with-reporting:
	kustomize build overlays/reporting | kubectl apply -f -

.PHONY: undeploy-base-with-reporting
undeploy-base-with-reporting:
	kustomize build overlays/reporting | kubectl delete -f -

