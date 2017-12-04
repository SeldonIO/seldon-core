notarget:
	@echo need target

run_jupyter_notebook:
	@jupyter-notebook --ip=127.0.0.1 --port=8888 --no-browser

run_core_builder:
	docker run --rm -it \
		-v $${HOME}/.docker:/root/.docker \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $${HOME}/.m2:/root/.m2 \
		-v $$(pwd):/work \
		seldonio/core-builder:0.1 bash


run_core_builder_in_minikube:
	eval $$(minikube docker-env) && \
		docker run --rm -it \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v /home/docker/.m2:/root/.m2 \
			-v $$(pwd):/work \
			seldonio/core-builder:0.1 bash

