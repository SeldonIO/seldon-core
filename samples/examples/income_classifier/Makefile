
infer.md: infer.ipynb
	jupyter nbconvert \
		infer.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output infer.md
	../../markdown-transforms.sh infer.md

train.md: train.ipynb
	jupyter nbconvert \
		train.ipynb \
		--ClearOutputPreprocessor.enabled=False \
		--to markdown \
		--output train.md
	../../markdown-transforms.sh train.md


.PHONY: upload_artifacts
upload_artifacts:
	gsutil cp -R classifier gs://seldon-models/scv2/examples/income
	gsutil cp -R explainer gs://seldon-models/scv2/examples/income
	gsutil cp -R drift-detector gs://seldon-models/scv2/examples/income
	gsutil cp -R outlier-detector gs://seldon-models/scv2/examples/income
	gsutil cp -R preprocessor gs://seldon-models/scv2/examples/income
	gsutil cp -R infer-data gs://seldon-models/scv2/examples/income

.PHONY: install-dev
install-dev:
	pip install -r requirements.txt
