
infer.md: infer.ipynb
	jupyter nbconvert \
		infer.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output infer.md
	../../markdown-transforms.sh infer.md
