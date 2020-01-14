
## For development if updating the Seldon Example Dashboard.

Save the dashboard to JSON by using the grafana API.
Find the dashboard using:

```
curl http://admin:password@localhost:3000/api/search?tag=seldon
```

Export by UID:

```
curl -v http://admin:password@localhost:3000/api/dashboards/uid/TcosMYEWz > dashboard.json
```

Manually edit to

 * Remove `uid` and `id`
 * set aliasColors to `{}` rather than `null`. TODO: fix this so we can make process more automated

A dashboard can be imported with:

```
curl -v http://admin:password@localhost:3000/api/dashboards/import -d "@dashboard.json" --header "Content-Type: application/json"
```

This API endpoint seems to be undocumented. The create dashboard referred to in docs did not work with exported dashboard. TODO: find out if we can use the documented API endpoint.

See [Grafana API docs](https://grafana.com/docs/grafana/latest/http_api/dashboard/)

