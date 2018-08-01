CORE_PYTHON_WAPPER_IMAGE=seldonio/core-python-wrapper:0.3

wrap_model:
	set -x && docker run --rm -it \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v "$$(pwd):/work" \
			$(CORE_PYTHON_WAPPER_IMAGE) \
			bash -c 'cd /work/models/mean_classifier && make -f makefile.ci wrap_model'

build_model_image:
	set -x && docker run --rm -it \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v "$$(pwd):/work" \
			$(CORE_PYTHON_WAPPER_IMAGE) \
			bash -c 'cd /work/models/mean_classifier && make -f makefile.ci build_model_image'


serve_model:
	set -x && docker run --rm -it --name=meanclassifier  gsunner/meanclassifier:v1_test

test_model:
	set -x && docker run --rm -it --link meanclassifier -v "$$(pwd):/work" $(CORE_PYTHON_WAPPER_IMAGE) bash -c 'python /wrappers/tester.py models/mean_classifier/contract.json $${MEANCLASSIFIER_PORT_5000_TCP_ADDR} $${MEANCLASSIFIER_PORT_5000_TCP_PORT} -p'

clean:
	rm -rfv ./models/mean_classifier/build


run_core_python_wrapper:
	docker run --rm -it \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v "$$(pwd):/work" \
			$(CORE_PYTHON_WAPPER_IMAGE) \
			bash

