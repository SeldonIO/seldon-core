

build_image:
	s2i build -E environment_grpc . seldon_python_openvino seldon-openvino-prediction


clean:
	rm -rf model
	rm -rf proto/__pycache__
	rm -f proto/*.py
	rm -r proto/*.proto
	rm -rf tensorflow
