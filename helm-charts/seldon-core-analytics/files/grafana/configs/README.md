
## For development if creating a new dashboard.

This helm chart deploys Grafana Dashboards using a sidecar
(https://github.com/helm/charts/tree/master/stable/grafana#sidecar-for-dashboards)

To modify or add a dashboard please follow the steps below.

  * Save the dashboard to JSON by exporting it and copy it to the `files/grafana/config` directory.
  * Save a new configmap in the `templates/grafana` directory for the dashboard using the template below, replacing <new-dashboard> with the name of your dashboard:
     * <new-dashboard>-configmap.yaml
     ```
     apiVersion: v1
     data:
     {{ (.Files.Glob "files/grafana/configs/<new-dashboard>.json").AsConfig | indent 2 }}
     kind: ConfigMap
     metadata:
       creationTimestamp: null
       name: <new-dashboard>
       namespace: {{ .Release.Namespace }}
       labels:
         seldon_dashboard: "1" 
     ```
