export ISTIO_VERSION=1.7.3
rm -rf istio-${ISTIO_VERSION}
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${ISTIO_VERSION} sh -
cd istio-${ISTIO_VERSION}/bin

#1.6 install steps were different
if  [[ $ISTIO_VERSION == 1.6* ]] ;
then

   echo 'Installing istio'
   ./istioctl manifest apply --set profile=default
   cat << EOF > ./local-cluster-gateway.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: empty
  components:
    ingressGateways:
      - name: cluster-local-gateway
        enabled: true
        label:
          istio: cluster-local-gateway
          app: cluster-local-gateway
        k8s:
          service:
            type: ClusterIP
            ports:
            - port: 15020
              name: status-port
            - port: 80
              name: http2
            - port: 443
              name: https
  values:
    gateways:
      istio-ingressgateway:
        debug: error
EOF

   ./istioctl manifest generate -f local-cluster-gateway.yaml > manifest.yaml
   kubectl apply -f manifest.yaml
   echo 'istio & gateway setup completed'


else

#1.7 (and hopefully above) steps
   echo 'Installing istio'
   cat << EOF > ./local-cluster-gateway.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  values:
    global:
      proxy:
        autoInject: disabled
      useMCP: false
  addonComponents:
    pilot:
      enabled: true
    prometheus:
      enabled: false
  components:
    ingressGateways:
      - name: cluster-local-gateway
        enabled: true
        label:
          istio: cluster-local-gateway
          app: cluster-local-gateway
        k8s:
          service:
            type: ClusterIP
            ports:
            - port: 15020
              name: status-port
            - port: 80
              targetPort: 8080
              name: http2
            - port: 443
              targetPort: 8443
              name: https
  values:
    gateways:
      istio-ingressgateway:
        debug: error
EOF

   ./istioctl install -f local-cluster-gateway.yaml
   echo 'istio & gateway setup completed'

fi