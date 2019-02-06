SHELL=/bin/bash

swagger.json:
	wget https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json

pod-template-spec.json: swagger.json
	python expand-validation.py > pod-template-spec.json

pod-template-spec-validation.tpl: pod-template-spec.json
	cat <(echo '{{- define "podTemplateSpec" }}') pod-template-spec.json <(echo '{{- end }}') > pod-template-spec-validation.tpl

hpa-spec.json: swagger.json
	python expand-validation.py --root "io.k8s.api.autoscaling.v2beta1.HorizontalPodAutoscalerSpec" > hpa-spec.json

clean:
	rm -f swagger.json
	rm -f pod-template-spec.json
	rm -f pod-template-spec-validation.tpl
	rm -f hpa-spec.json
