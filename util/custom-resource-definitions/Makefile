SHELL=/bin/bash

swagger.json:
	wget https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json

pod-spec.json: swagger.json
	python expand-validation.py --root "io.k8s.api.core.v1.PodSpec" > pod-spec.json

container.json: swagger.json
	python expand-validation.py --root "io.k8s.api.core.v1.Container" > container.json

hpa-spec.json: swagger.json
	python expand-validation.py --root "io.k8s.api.autoscaling.v2beta1.MetricSpec" > hpa-spec.json

object-meta.json: swagger.json
	python expand-validation.py --root "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta" > object-meta.json

crd.yaml: object-meta.json hpa-spec.json pod-spec.json container.json
	python create-crd.py > crd.yaml

.PHONY: clean
clean:
	rm -f swagger.json
	rm -f pod-spec.json
	rm -f hpa-spec.json
	rm -f container.json
	rm -f object-meta.json
	rm -f crd.yaml
