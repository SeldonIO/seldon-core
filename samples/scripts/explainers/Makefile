VERSION := $(shell grep seldonio/mlserver ../../../scheduler/Makefile | cut -d':' -f2)
MLSERVER_VERSION ?= ${VERSION}
MLSERVER_FOLDER = mlserver_${MLSERVER_VERSION}
PYTHON_SCRIPT = make_test_models.py
OUTPUT_DIR ?= models

define generate_file
    mkdir -p ${OUTPUT_DIR}/${1}/data
	.env/bin/python3 ${PYTHON_SCRIPT} --model ${1} --model_dir ${OUTPUT_DIR}/${1}/data
	REPLACE_ME=${1} envsubst < model-settings.json.template > ${OUTPUT_DIR}/${1}/model-settings.json
endef

.PHONY: train-all
train-all: kernel_shap anchor_text_moviesentiment

.PHONY: upload-all
upload-all: upload_kernel_shap upload_anchor_text_moviesentiment

.PHONY: env
env:
	python3 -m venv .env
	.env/bin/pip install \
		mlserver==${MLSERVER_VERSION} \
		mlserver-sklearn==${MLSERVER_VERSION} \
		mlserver-alibi-explain==${MLSERVER_VERSION} \
		mlserver-xgboost==${MLSERVER_VERSION}

.PHONY: anchor_image
anchor_image: env
	$(call generate_file,anchor_image)

.PHONY: kernel_shap
kernel_shap: env
	 $(call generate_file,kernel_shap)

.PHONY: tree_shap
tree_shap: env
	 $(call generate_file,tree_shap)

.PHONY: ale
ale: env
	 $(call generate_file,ale)

.PHONY: anchor_tabular
anchor_tabular: env
	 $(call generate_file,anchor_tabular)

.PHONY: anchor_tabular_income
anchor_tabular_income: env
	 $(call generate_file,anchor_tabular_income)

.PHONY: anchor_text_moviesentiment
anchor_text_moviesentiment: env
	 $(call generate_file,anchor_text_moviesentiment)


upload_anchor_text_moviesentiment:
	gsutil cp -R models/anchor_text_moviesentiment gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/moviesentiment-sklearn-explainer

upload_kernel_shap:
	gsutil cp -R models/kernel_shap gs://seldon-models/scv2/samples/${MLSERVER_FOLDER}/wine-kernel-shap-explainer
