
infer_local.md: infer_local.ipynb
	jupyter nbconvert \
		infer_local.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output infer_local.md

infer.md: infer.ipynb
	jupyter nbconvert \
		infer.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output infer.md


.PHONY: upload_artifacts
upload_artifacts:
	gsutil cp -R pandasquery gs://seldon-models/scv2/examples/pandasquery

.PHONY: install-dev
install-dev:
	pip install -r requirements.txt
