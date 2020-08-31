

serving_git:
	git clone -b r1.14 https://github.com/tensorflow/serving.git serving_git


tensorflow_git:
	git clone -b r1.15 https://github.com/tensorflow/tensorflow.git tensorflow_git


create_protos: serving_git tensorflow_git
	cd serving_git && python -m grpc.tools.protoc ./tensorflow_serving/apis/*.proto --python_out=.. --grpc_python_out=.. --proto_path=. --proto_path=../tensorflow_git
	cd tensorflow_git && python -m grpc.tools.protoc ./tensorflow/core/protobuf/*.proto --python_out=.. --grpc_python_out=.. --proto_path=.
	cd tensorflow_git && python -m grpc.tools.protoc ./tensorflow/core/example/*.proto --python_out=.. --grpc_python_out=.. --proto_path=.
	cd tensorflow_git && python -m grpc.tools.protoc ./tensorflow/core/framework/*.proto --python_out=.. --grpc_python_out=.. --proto_path=.
	touch tensorflow/core/protobuf/__init__.py
	touch tensorflow/core/example/__init__.py
	touch tensorflow/core/framework/__init__.py
	touch tensorflow_serving/apis/__init__.py



clean:
	rm -rf serving_git
	rm -rf tensorflow_git
