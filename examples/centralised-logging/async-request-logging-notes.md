
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
See https://github.com/knative/eventing/tree/master/contrib/kafka/config

Install eventing with https://knative.dev/v0.3-docs/eventing/
May need to apply scripts twice due to https://github.com/knative/eventing/issues/680

See example:
See https://github.com/knative/docs/tree/master/docs/eventing/samples/container-source

For ContainerSource example do
`kubectl apply -f https://raw.githubusercontent.com/knative/docs/master/docs/eventing/samples/container-source/service.yaml`
`kubectl apply -f https://raw.githubusercontent.com/knative/docs/master/docs/eventing/samples/container-source/heartbeats-source.yaml`

Seems like what we want to do is have a simple source that takes a raw http request and transforms to CloudEvents format.
Actually don't want to do that in the engine anyway.
So something like in https://github.com/knative/eventing-contrib/blob/master/cmd/heartbeats/main.go
Then need a sink that transforms and logs to stdout. So quite similar to the heartbeats event-display service.
But we need to have a way to send from something that isn't a Source CR.
Looks like broker-trigger is the way forward. https://github.com/knative/eventing-contrib/issues/453#issuecomment-500597056

Engine will need to enrich request with both puid and the pod, container and request it pertains to. Will need to use https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/