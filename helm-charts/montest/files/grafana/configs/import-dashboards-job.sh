
GRAFANA_USER=admin
GRAFANA_PASS=${GF_SECURITY_ADMIN_PASSWORD}
HOST=grafana-prom
PORT=80

# For DEV outside of cluster
if [ ! -e /var/run/secrets/kubernetes.io/serviceaccount ]; then
    HOST=localhost
    PORT=3000
fi
echo "Using ${HOST}:${PORT}"

recreate_datasource() {
    curl --silent --fail --show-error --request DELETE http://${GRAFANA_USER}:${GRAFANA_PASS}@${HOST}:${PORT}/api/datasources/name/prometheus

    curl --silent --fail --show-error --request POST http://${GRAFANA_USER}:${GRAFANA_PASS}@${HOST}:${PORT}/api/datasources --header "Content-Type: application/json" --data-binary "@prometheus-datasource.json"
}

add_dashboard() {
    file=$1
    SRC_NAME=$2
    echo ""
    echo "importing $file"

    (   echo '{"dashboard":';
        cat "$file";
        echo ',"overwrite":true,"inputs":[{"name":"'${SRC_NAME}'","type":"datasource","pluginId":"prometheus","value":"prometheus"}]}'
    ) | jq -c '.' | curl --silent --fail --show-error --request POST http://${GRAFANA_USER}:${GRAFANA_PASS}@${HOST}:${PORT}/api/dashboards/import --header "Content-Type: application/json" --data-binary "@-"

}

recreate_datasource
add_dashboard grafana-net-2115-dashboard.json DS_PROM
add_dashboard metrics-test-dashboard.json DS_PROM
add_dashboard predictions-analytics-dashboard.json DS_PROM

