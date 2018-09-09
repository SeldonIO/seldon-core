
## For development if updating the Seldon Example Dashboard.

  * Save the dashboard to JSON by exporting it.
  * Overwrite the prediction-analytics-dashboard.json with the exported JSON
  * Run ```./convert-exported-graph.sh```
  * Then import the new dashboard
     * Port forward the grafana port
     ```
     kubectl port-forward $(kubectl get pods -n seldon -l app=grafana-prom-server -o jsonpath='{.items[0].metadata.name}') -n seldon 3000:3000
     ```
     * export the password used when starting the analytics, e.g.
     ```
     export GF_SECURITY_ADMIN_PASSWORD=password
     ```
     * run ```import-dashboards-job.sh```

