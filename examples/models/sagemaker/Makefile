

train_local:
	SM_OUTPUT_DATA_DIR="X" SM_MODEL_DIR="X" SM_CHANNEL_TRAIN="X" python scikit_learn_iris.py --model-dir model --train data


build_image:
	docker build -t seldonio/sagemaker:0.1 .

serve_local:
	#docker run -it --rm -p 8080:8080 -e SAGEMAKER_SERVING_MODULE="serving:main" -e SAGEMAKER_PROGRAM="scikit_learn_iris.py" -e SAGEMAKER_SUBMIT_DIR="file:///module" seldonio/sagemaker:0.1 serve
	docker run -it --rm -p 8080:8080 -e SAGEMAKER_SERVING_MODULE="serving:main" -e SAGEMAKER_PROGRAM="scikit_learn_iris.py" seldonio/sagemaker:0.1 serve

