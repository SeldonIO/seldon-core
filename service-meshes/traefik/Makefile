
README.md:
	jupyter nbconvert \
		traefik.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output README.md


static/single-model.yaml: ./config/single-model/model.yaml ./config/single-model/ingressroute.yaml ./config/single-model/iris-header-middleware.yaml
	kustomize build config/single-model > static/single-model.yaml

