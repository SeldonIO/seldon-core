
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
