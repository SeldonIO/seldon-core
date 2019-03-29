SC_KSONNET_FOLDER=../../seldon-core/seldon-core/json
SCA_KSONNET_FOLDER=../../seldon-core/seldon-core-analytics/json

.PHONY: start-minikube
start-minikube:
	minikube start --vm-driver kvm2 --memory 4096 --kubernetes-version v1.8.6 --feature-gates=CustomResourceValidation=true --extra-config=apiserver.authorization-mode=RBAC

kubectl:
	curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.8.6/bin/linux/amd64/kubectl
	chmod +x ./kubectl

install_kubectl:
	sudo mv ./kubectl /usr/local/bin/kubectl

template_0.2_single_namespace.yaml template_0.2_cluster_wide.yaml:
	helm template --set ambassador.enabled=true --set single_namespace=true ../../helm-charts/seldon-core > template_0.2_single_namespace.yaml
	helm template --set ambassador.enabled=true --set single_namespace=false --set ambassador.rbac.namespaced=false --set ambassador.env.AMBASSADOR_SINGLE_NAMESPACE=false ../../helm-charts/seldon-core > template_0.2_cluster_wide.yaml

template_0.2_single_namespace.json template_0.2_cluster_wide.json: template_0.2_single_namespace.yaml template_0.2_cluster_wide.yaml
	python y2j.py < template_0.2_single_namespace.yaml > template_0.2_single_namespace.json
	python y2j.py < template_0.2_cluster_wide.yaml > template_0.2_cluster_wide.json


seldon-core-release-0.1:
	git clone --branch release-0.1 https://github.com/SeldonIO/seldon-core.git seldon-core-release-0.1

template_0.1.yaml: seldon-core-release-0.1
	helm template --set ambassador.enabled=true seldon-core-release-0.1/helm-charts/seldon-core > template_0.1.yaml

template_0.1.json: template_0.1.yaml kubectl
	./kubectl convert -f template_0.1.yaml -o json > template_0.1.json


template_analytics.txt:
	helm template ../../helm-charts/seldon-core-analytics > template_analytics.txt

analytics.json: template_analytics.txt
	python yj2j.py template_analytics.txt > analytics.json

######

.PHONY: build
build: template_0.2_single_namespace.json template_0.2_cluster_wide.json template_0.1.json

.PHONY: release_analytics
release_analytics:
	cp analytics.json ${SCA_KSONNET_FOLDER}

.PHONY: release
release:
	cp template_0.2_single_namespace.json ${SC_KSONNET_FOLDER}
	cp template_0.2_cluster_wide.json ${SC_KSONNET_FOLDER}
	cp template_0.1.json ${SC_KSONNET_FOLDER}

.PHONY: clean
clean:
	rm -f *.json
	rm -f *.yaml
	rm -rf seldon-core-release-0.1
	rm -f kubectl
	rm -f template_analytics.txt
	rm -f analytics.json

