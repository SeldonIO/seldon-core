cd ../../helm-charts

echo "Creating Weave Pods"

kubectl create clusterrolebinding mf --clusterrole=cluster-admin --user="922480082290-compute@developer.gserviceaccount.com"
kubectl apply -n kube-system -f "https://cloud.weave.works/k8s.yaml?service-token=eu5jdkanx8th8htg9w7r14b4i6ytgz6q&k8s-version=$(kubectl version | base64 | tr -d '\n')"

sleep 5

helm init

echo "Waiting for helm initialisation"
until [ $(kubectl get pods -n kube-system -l app=helm,name=tiller -o jsonpath="{.items[0].status.containerStatuses[0].ready}") = "true" ]
do
    echo "."
    sleep 2
done

helm install seldon-core --name seldon-core \
        --set cluster_manager_client_secret=mysecret \
        --set grafana_prom_admin_password=mypassword \
        --set persistence.enabled=false \
        --set cluster_manager_service_type=LoadBalancer \
	--set cluster_manager.image.tag=0.2.16_mab_demo_stable \
        --set grafana_prom_service_type=LoadBalancer \
        --set apife_service_type=LoadBalancer

echo "Adding repo secret"
cd ../../seldon-deploy/gcr
./create-registry-secret

echo "Waiting for loadbalancers"
cd ../../seldon-core/tests

until [[ $(kubectl get svc seldon-cluster-manager -o jsonpath="{.status.loadBalancer.ingress[0].ip}") ]]
do
    echo "."
    sleep 5
done
until [[ $(kubectl get svc seldon-apiserver -o jsonpath="{.status.loadBalancer.ingress[0].ip}") ]]
do
    echo "."
    sleep 5
done
make cm_update_endpoint_for_loadbalancer
make api_update_endpoint_for_loadbalancer

until [[ $(kubectl get svc grafana-prom -o jsonpath="{.status.loadBalancer.ingress[0].ip}") ]]
do
    echo "."
    sleep 5
done

xdg-open http://$(kubectl get svc grafana-prom -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
xdg-open https://cloud.weave.works/instances

echo "Starting Jupyter"
cd demos
jupyter notebook MAB\ Demo.ipynb
