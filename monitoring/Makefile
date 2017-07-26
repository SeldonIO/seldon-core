show_all:
	watch 'kubectl get all -o wide --show-labels --all-namespaces'
show_default:
	watch 'kubectl get all -o wide --show-labels --namespace=default'

show_pv:
	kubectl get pv -o wide --show-labels



port_forward_prometheus:
	POD_NAME=$$(kubectl --namespace default get pod -l app=prometheus-server -o template --template="{{(index .items 0).metadata.name}}") && \
        kubectl port-forward $${POD_NAME} 9090:9090

port_forward_grafana:
	POD_NAME=$$(kubectl --namespace default get pod -l app=grafana-prom-server -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl port-forward $${POD_NAME} 3000:3000

port_forward_alertmanager:
	POD_NAME=$$(kubectl --namespace default get pod -l app=alertmanager-server -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl port-forward $${POD_NAME} 9093:9093


get_current_context:
	@kubectl config current-context

set_context_minikube_namespace_default:
	@kubectl config set-context minikube --namespace=default
	@kubectl config use-context minikube
	@kubectl config get-contexts minikube

clean:
	@rm -fv \
		alertmanager/_alertmanager-deployment.json \
		alertmanager/_configmap.yaml \
		grafana/_configmap.yaml \
		prometheus/_prometheus-deployment.json \
		prometheus/_prometheus-rules-configmap.yaml \
		prometheus/_prometheus-server-conf-configmap.yaml

