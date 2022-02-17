

build_image:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python36:1.14.0-dev seldon-resnet2.4


clean:
	rm -rf model
	rm -rf proto/__pycache__
	rm -f proto/*.py
	rm -r proto/*.proto
	rm -rf tensorflow
