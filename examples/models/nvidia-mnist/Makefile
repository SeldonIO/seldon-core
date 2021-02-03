TRANSFORMER_IMAGE=seldonio/mnist-caffe2-transformer:0.1

clean:
	rm -f rm -f tensorrt_mnist/1/model.plan
	rm -rf MNIST_data
	rm -f mnist.json
	rm -f tmp.json

build_transformer:
	s2i build . seldonio/seldon-core-s2i-python3:1.2.4 ${TRANSFORMER_IMAGE}

push_transformer:
	docker push ${TRANSFORMER_IMAGE}

