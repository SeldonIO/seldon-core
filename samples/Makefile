
local-examples.md: local-examples.ipynb
	jupyter nbconvert \
		local-examples.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output local-examples.md
	./markdown-transforms.sh local-examples.md

k8s-examples.md: k8s-examples.ipynb
	jupyter nbconvert \
		k8s-examples.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output k8s-examples.md
	./markdown-transforms.sh k8s-examples.md

pipeline-examples.md: pipeline-examples.ipynb
	jupyter nbconvert \
		pipeline-examples.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output pipeline-examples.md
	./markdown-transforms.sh pipeline-examples.md

explainer-examples.md: explainer-examples.ipynb
	jupyter nbconvert \
		explainer-examples.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output explainer-examples.md
	./markdown-transforms.sh explainer-examples.md

.PHONY: create
create: local-examples.md k8s-examples.md pipeline-examples.md explainer-examples.md
