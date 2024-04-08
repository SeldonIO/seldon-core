// adapted from https://github.com/grafana/xk6-kubernetes/blob/main/examples/deployment_operations.js
// note that xk6 needs to be installed to run this script
// also note that we need to have a secret for kubectl to work with the cluster
// run `make create-scecret-kube` to create the secret based on the locally available kubeconfig
import { Kubernetes } from "k6/x/kubernetes";
import { describe, expect } from "https://jslib.k6.io/k6chaijs/4.3.4.3/index.js";
import { load, dump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
import { getConfig } from '../components/settings.js';

let yaml = `
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple1
  namespace: ${getConfig().namespace}
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
`

export default function () {
    const kubernetes = new Kubernetes();

    describe('YAML-based resources', () => {
        let yamlObject = load(yaml)
        const name = yamlObject.metadata.name
        const ns = yamlObject.metadata.namespace

        describe('Create our Model using the YAML definition', () => {
            kubernetes.apply(yaml)
            let created = kubernetes.get("Model.mlops.seldon.io", name, ns)
            expect(created.metadata, 'new Model').to.have.property('uid')
        })

        describe('Update our Model with a modified YAML definition', () => {
            const newValue = 2
            yamlObject.spec.replicas = newValue
            let newYaml = dump(yamlObject)

            kubernetes.apply(newYaml)
            let updated = kubernetes.get("Model.mlops.seldon.io", name, ns)
            expect(updated.spec.replicas, 'changed value').to.be.equal(newValue)
        })

        describe('Remove our Model to cleanup', () => {
            kubernetes.delete("Model.mlops.seldon.io", name, ns)
        })
    })

}
