# End to End Tests

We use [Kind](https://github.com/kubernetes-sigs/kind) to run our tests.

To get everything setup run:

```
kind_test_setup.sh
```

Activate kind kubernetes config:

```
export KUBECONFIG="$(kind get kubeconfig-path)"
kindworkerid=$(docker ps -f "name=kind-worker" -q)
kindip=$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" $kindworkerid)
ambport=$(kubectl get svc -n seldon ambassador -o jsonpath={.spec.ports[0].nodePort})
export API_AMBASSADOR="$kindip:$ambport"
```

Then to run the tests:

```
make test
```


