BUCKET=seldon-sagemaker-testing

scikit_learn_iris_code.tar.gz:
	cd scikit_learn_iris && tar -cvf ../scikit_learn_iris_code.tar .
	gzip -f scikit_learn_iris_code.tar

.PHONY: upload_scikit_code
upload_scikit_code:
	aws s3 cp scikit_learn_iris_code.tar.gz s3://${BUCKET}/scikit_learn_iris_code.tar.gz


#Need to build sklearn image

docker_serve:
	docker run -it --rm -p 8080:8080 -e SAGEMAKER_MODEL_DIRECTORY="s3://seldon-sagemaker-testing/scikit_learn_iris/sagemaker-scikit-learn-2019-01-04-19-26-40-470/output/model.tar.gz" -e SAGEMAKER_SUBMIT_DIRECTORY="s3://seldon-sagemaker-testing/scikit_learn_iris_code.tar.gz"  -e SAGEMAKER_PROGRAM="scikit_learn_iris.py" -v ~/.aws:/root/.aws sklearn-final:0.20.0-cpu-py3 serve


clean:
	rm -f scikit_learn_iris_code.tar.gz


