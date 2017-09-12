To run iwth istio:
  Prepare: if running on minikube you will need a larger than default cluster. Tested with 9g but should work with less:
  	   minikube start --memory=9000 --disk-size "40g"
  1. Install istio on cluster using Helm.  See: https://github.com/kubernetes/charts/tree/master/incubator/istio
     a) One time only:
     	helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
     b) helm install incubator/istio --name istio-test --set auth.enabled=false
  2. Change cluster-manger settings.sh to have:
       SPRING_OPTS=--io.seldon.clustermanager.istio-enabled=true
     This allows the cluster-manager to install envoy sidecar in new deployments
  3. (Needs fixing) Change start-all to use 
       cd "${STARTUP_DIR}/api-frontend" && make start_apife_istio
     This uses a kubernetes json with Envoy sidecar next to api front end

To remove istio:
   helm delete istio-test --purge

-----------------------

Verify all Pods are running

  kubectl get pods --namespace default

Verifying the Grafana dashboard

  export POD_NAME=$(kubectl get pods --namespace default -l "component=istio-test-istio-grafana" -o jsonpath="{.items[0].metadata.name}")
  kubectl port-forward $POD_NAME 3000:3000 --namespace default
  echo http://127.0.0.1:3000/dashboard/db/istio-dashboard

Verifying the ServiceGraph service

  export POD_NAME=$(kubectl get pods --namespace default -l "component=istio-test-istio-servicegraph" -o jsonpath="{.items[0].metadata.name}")
  kubectl port-forward $POD_NAME 8088:8088 --namespace default
  echo http://127.0.0.1:8088/dotviz

Deploy your App!

  kubectl create -f <(istioctl kube-inject -f <your-app-spec>.yaml)

Or deploy the BookInfo App!

  https://istio.io/docs/samples/bookinfo.html

Using Istioctl

  istioctl [command]


Also:

Verify Zipkin service:

	kubectl port-forward $(kubectl get pod -l "component=istio-test-istio-zipkin" -o jsonpath='{.items[0].metadata.name}') 9411:9411
    	echo http://127.0.0.1:9411
