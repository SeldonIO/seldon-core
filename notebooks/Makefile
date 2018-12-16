install_python_protos: update_protos build_protos

update_protos:
	cd ../proto/tensorflow && make create_protos
	cp -vr ../proto/tensorflow/tensorflow scripts
	cp ../proto/prediction.proto ./proto

build_protos: update_protos
	python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. ./proto/prediction.proto


clean:
	rm -f proto/prediction_pb2_grpc.py  
	rm -f proto/prediction_pb2.py
	rm -f proto/prediction.proto
