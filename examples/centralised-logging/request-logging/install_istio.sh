# Based on https://knative.dev/docs/install/installing-istio/

# Download and unpack Istio
#export ISTIO_VERSION=1.3.6
export ISTIO_VERSION=1.1.6
rm -rf istio-${ISTIO_VERSION}
curl -L https://git.io/getLatestIstio | sh -
cd istio-${ISTIO_VERSION}

for i in install/kubernetes/helm/istio-init/files/crd*yaml; do kubectl apply -f $i; done

echo "Sleep while CRDs become ready"
sleep 20

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: istio-system
  labels:
    istio-injection: disabled
EOF

# A lighter template, with just pilot/gateway.
# Based on install/kubernetes/helm/istio/values-istio-minimal.yaml
helm template --namespace=istio-system \
     --set prometheus.enabled=false \
     --set mixer.enabled=false \
     --set mixer.policy.enabled=false \
     --set mixer.telemetry.enabled=false \
     `# Pilot doesn't need a sidecar.` \
     --set pilot.sidecar=false \
     --set pilot.resources.requests.memory=128Mi \
     `# Disable galley (and things requiring galley).` \
     --set galley.enabled=false \
     --set global.useMCP=false \
     `# Disable security / policy.` \
     --set security.enabled=false \
     --set global.disablePolicyChecks=true \
     `# Disable sidecar injection.` \
     --set sidecarInjectorWebhook.enabled=false \
     --set global.proxy.autoInject=disabled \
     --set global.omitSidecarInjectorConfigMap=true \
     --set gateways.istio-ingressgateway.autoscaleMin=1 \
     --set gateways.istio-ingressgateway.autoscaleMax=2 \
     `# Set pilot trace sampling to 100%` \
     --set pilot.traceSampling=100 \
     --set global.mtls.auto=false \
     install/kubernetes/helm/istio \
     > ./istio-lean.yaml

kubectl apply -f istio-lean.yaml


# Add the extra gateway.
helm template --namespace=istio-system \
     --set gateways.custom-gateway.autoscaleMin=1 \
     --set gateways.custom-gateway.autoscaleMax=2 \
     --set gateways.custom-gateway.cpu.targetAverageUtilization=60 \
     --set gateways.custom-gateway.labels.app='cluster-local-gateway' \
     --set gateways.custom-gateway.labels.istio='cluster-local-gateway' \
     --set gateways.custom-gateway.type='ClusterIP' \
     --set gateways.istio-ingressgateway.enabled=false \
     --set gateways.istio-egressgateway.enabled=false \
     --set gateways.istio-ilbgateway.enabled=false \
     --set global.mtls.auto=false \
     install/kubernetes/helm/istio \
     -f install/kubernetes/helm/istio/example-values/values-istio-gateways.yaml \
    | sed -e "s/custom-gateway/cluster-local-gateway/g" -e "s/customgateway/clusterlocalgateway/g" \
	  > ./istio-local-gateway.yaml

kubectl apply -f istio-local-gateway.yaml

