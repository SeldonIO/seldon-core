SHELL=/bin/bash

swagger.json:
	wget https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json

pod-spec.json: swagger.json
	python expand-validation.py --root "io.k8s.api.core.v1.PodSpec" > pod-spec.json

pod-spec-validation.tpl: pod-spec.json
	cat <(echo '{{- define "podSpec" }}') pod-spec.json <(echo '{{- end }}') > pod-spec-validation.tpl

hpa-spec.json: swagger.json
	python expand-validation.py --root "io.k8s.api.autoscaling.v2beta1.MetricSpec" > hpa-spec.json

hpa-spec-validation.tpl: hpa-spec.json
	cat <(echo '{{- define "hpaSpec" }}') hpa-spec.json <(echo '{{- end }}') > hpa-spec-validation.tpl

object-meta.json: swagger.json
	python expand-validation.py --root "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta" > object-meta.json

object-meta-validation.tpl: object-meta.json
	cat <(echo '{{- define "objectMeta" }}') object-meta.json <(echo '{{- end }}') > object-meta-validation.tpl

.PHONY: all
all: object-meta-validation.tpl hpa-spec-validation.tpl pod-spec-validation.tpl

.PHONY: update-helm
update-helm: all
	cp object-meta-validation.tpl ../../helm-charts/seldon-core-controller/templates/_object-meta-validation.tpl
	cp pod-spec-validation.tpl ../../helm-charts/seldon-core-controller/templates/_pod-spec-validation.tpl
	cp hpa-spec-validation.tpl ../../helm-charts/seldon-core-controller/templates/_hpa-spec-validation.tpl

update-ksonnet: all
	cp object-meta.json ../../seldon-core/seldon-core/json/object-meta-validation.json
	cp hpa-spec.json ../../seldon-core/seldon-core/json/hpa-validation.json
	cp pod-spec.json ../../seldon-core/seldon-core/json/pod-spec-validation.json

.PHONY: clean
clean:
	rm -f swagger.json
	rm -f pod-spec.json
	rm -f pod-spec-validation.tpl
	rm -f hpa-spec.json
	rm -f hpa-spec-validation.tpl
	rm -f object-meta.json
	rm -f object-meta-validation.tpl
