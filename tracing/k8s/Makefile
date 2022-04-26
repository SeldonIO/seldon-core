# https://cert-manager.io/docs/installation/#default-static-install
.PHONY: install-cert-manager
install-cert-manager:
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
	kubectl wait --for condition=ready --timeout=300s pods --all -n cert-manager

# https://cert-manager.io/docs/installation/#default-static-install
.PHONY: uninstall-cert-manager
uninstall-cert-manager:
	kubectl delete --ignore-not-found=true -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml

# https://www.jaegertracing.io/docs/1.33/operator/
.PHONY:install-jaeger-operator
install-jaeger-operator: 
	kubectl create namespace observability || echo "observability namespace exists"
	kubectl create -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.33.0/jaeger-operator.yaml -n observability
	kubectl wait --for condition=established --timeout=60s crd/jaegers.jaegertracing.io
	kubectl wait --for condition=ready --timeout=300s pods --all -n observability

.PHONY: install-jaeger
install-jaeger:
	kubectl apply -f jaeger.yaml -n seldon-mesh

.PHONY: uninstall-jaeger
uninstall-jaeger:
	kubectl delete -f jaeger.yaml -n seldon-mesh

.PHONY:uninstall-jaeger-operator
uninstall-jaeger-operator:
	kubectl delete --ignore-not-found=true -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.33.0/jaeger-operator.yaml -n observability
	kubectl delete ns observability

.PHONY:install-opentelemetry-operator
install-opentelemetry-operator:
	kubectl apply -f https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/opentelemetry-operator.yaml
	kubectl wait --for condition=ready --timeout=300s pods --all -n opentelemetry-operator-system

.PHONY:uninstall-opentelemetry-operator
uninstall-opentelemetry-operator:
	kubectl delete --ignore-not-found=true -f https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/opentelemetry-operator.yaml

.PHONY:install-opentelemetry-collector
install-opentelemetry-collector:
	kubectl create -f otel-collector.yaml

.PHONY:uninstall-opentelemetry-collector
uninstall-opentelemetry-collector:
	kubectl delete -f otel-collector.yaml

.PHONY: install-tracing
install-tracing: install-cert-manager install-jaeger-operator install-jaeger install-opentelemetry-operator install-opentelemetry-collector

.PHONY: uninstall-tracing
uninstall-tracing: uninstall-opentelemetry-collector uninstall-jaeger uninstall-jaeger-operator uninstall-opentelemetry-operator uninstall-cert-manager
