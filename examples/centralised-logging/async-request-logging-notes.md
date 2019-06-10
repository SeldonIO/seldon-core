
Create minikube cluster with knative recommendations for resource - https://knative.dev/v0.3-docs/install/knative-with-minikube/

Follow the EFK minikube setup from README.md in this dir.

Run knative-setup.sh

Run `kubectl apply -f knative-example-service.yaml`

Then:

```
export IP_ADDRESS=$(minikube ip):$(kubectl get svc istio-ingressgateway --namespace istio-system --output 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')
curl -H "Host: helloworld-go.default.example.com" http://${IP_ADDRESS}
curl -H "Host: helloworld-go.default.example.com" http://${IP_ADDRESS}
```

In kibana filter by `kubernetes.container_name` value of `user-container`. You'll see the requests.

TODO: should we filter out the istio-proxy, istio-init and queue-proxy containers? They log much more than the user-container. Fluentd filtering might be:

```
<filter kubernetes.**>
    @type grep
    <exclude>
        key $.kubernetes.container_name
        pattern istio-proxy
    </exclude>
</filter>
```

TODO: Here we invoke the service via HTTP. The first curl is a bit slow so seems to sync (waits for service to warm up).
Can we use eventing to bring in a broker and call async? Then we could use that to forward requests to be logged to a logging component.