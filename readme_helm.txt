Using Seldon Core Helm chart

1. Install Helm

    https://github.com/kubernetes/helm


2. Initialize helm

    $ helm init


3. Launch Seldon Core

    $ cd seldon-core/helm-charts
    $ helm install seldon-core --name seldon-core \
        --set cluster_manager_client_secret=<your-cluster-manager-secret> \
        --set grafana_prom_admin_password=<your-grafana-prom-password> \
        --set persistence.enabled=false


4. Stop/Delete Seldon Core

    $ cd seldon-core/helm-charts
    $ helm delete seldon-core --purge


5. Using persistence

   Use the scripts in seldon-core/persistence to create a volume.
   Then launch Seldon Core with "--set persistence.enabled=true"


6. Example of other options for development

    $ cd seldon-core/helm-charts
    $ helm install seldon-core --name seldon-core \
        --set cluster_manager_client_secret=your-cluster-manager-secret> \
        --set cluster_manager_service_type=NodePort \
        --set grafana_prom_service_type=NodePort \
        --set apife_service_type=NodePort \
        --set cluster_manager.image.tag=0.2.14_metrics \
        --set apife.image.tag=0.0.5_metrics \
        --set spring_opts="--io.seldon.clustermanager.engine-container-image-and-version=seldonio/engine:0.1.5_metrics_v2 --io.seldon.clustermanager.istio-enabled=true"

