VERSION := $(shell cat ../../../../version.txt)

all: env train collect upload

model: env train

env:
	python3 -m venv .env
	./.env/bin/pip install --upgrade pip setuptools
	./.env/bin/pip install -r requirements.txt

train:
	.env/bin/mlflow run . -P alpha=0.5 -P l1_ratio=0.5

collect:
	cp metadata.yaml output/
	cd mlruns && find . -name "MLmodel" -type f -exec cp {} ../output \;
	cd mlruns && find . -name "conda.yaml" -type f -exec cp {} ../output \;
	cd mlruns && find . -name "model.pkl" -type f -exec cp {} ../output \;
	cd output && \
		conda env create --force -n elasticnet-wine-env -f conda.yaml && \
		conda-pack -n elasticnet-wine-env -f -o environment.tar.gz

upload:
	gsutil -m cp output/conda.yaml gs://seldon-models/v${VERSION}/mlflow/elasticnet_wine/
	gsutil -m cp output/metadata.yaml gs://seldon-models/v${VERSION}/mlflow/elasticnet_wine/
	gsutil -m cp output/MLmodel gs://seldon-models/v${VERSION}/mlflow/elasticnet_wine/
	gsutil -m cp output/model.pkl gs://seldon-models/v${VERSION}/mlflow/elasticnet_wine/
	gsutil -m cp output/environment.tar.gz gs://seldon-models/v${VERSION}/mlflow/elasticnet_wine/
