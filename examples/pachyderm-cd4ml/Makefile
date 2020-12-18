build:
	make -C RSIServing build
	make -C RSITraining build
	make -C RSIDeployer build

push: build
	make -C RSIServing push
	make -C RSITraining push
	make -C RSIDeployer push

kind_load: build
	make -C RSIServing kind_load
	make -C RSITraining kind_load
	make -C RSIDeployer kind_load