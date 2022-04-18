
README.md:
	jupyter nbconvert \
		istio.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output README.md


static/single-model.yaml: ./config/single-model/model.yaml ./config/single-model/virtualservice.yaml
	kustomize build config/single-model > static/single-model.yaml

static/traffic-split.yaml: ./config/traffic-split/models.yaml ./config/single-model/virtualservice.yaml
	kustomize build config/traffic-split > static/traffic-split.yaml
