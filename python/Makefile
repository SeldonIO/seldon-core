SELDON_CORE_DIR=..
VERSION=0.3.0

.PHONY: get_apis
get_apis:
	# Protobuf
	cp ${SELDON_CORE_DIR}/proto/prediction.proto seldon_core/proto/
	$(MAKE) -C ../proto/tensorflow/ create_protos
	cp -r $(SELDON_CORE_DIR)/proto/tensorflow/tensorflow seldon_core/
	$(MAKE) -C ../proto/tensorflow clean

	# OpenAPI
	cp $(SELDON_CORE_DIR)/openapi/wrapper.oas3.json seldon_core/openapi/seldon.json

.PHONY: build_apis
build_apis: get_apis
	# Protobuf
	cd seldon_core && python -m grpc.tools.protoc -I./ --python_out=./ --grpc_python_out=./ --mypy_out=./ ./proto/prediction.proto
	sed -i "s/from proto/from seldon_core.proto/g" seldon_core/proto/prediction_pb2_grpc.py

.PHONY: update_version
update_version:
	sed -i "s/__version__ = .*/__version__ = '$(VERSION)'/g" seldon_core/version.py

.PHONY: update_package
update_package: get_apis build_apis update_version

.PHONY: install
install:
	pip install -e .

.PHONY: install-dev
install-dev:
	pip install -e . -r requirements.txt

.PHONY: uninstall
uninstall:
	pip uninstall seldon-core

.PHONY: test
test:
	python setup.py test

.PHONY: type_check
type_check:
	mypy --ignore-missing-imports --no-strict-optional seldon_core

.PHONY: build_pypi
build_pypi:
	python setup.py sdist bdist_wheel

.PHONY: build_conda
build_conda:
	mkdir -p conda-bld
	conda build conda.recipe/meta.yaml -c conda-forge --output-folder conda-bld

.PHONY: push_pypi_test
push_pypi_test:
	twine upload --repository-url https://test.pypi.org/legacy/ dist/*

.PHONY: push_pypi
push_pypi:
	twine upload dist/*

.PHONY: push_conda
push_conda:
	@echo "1st push: need to submit a pull-request to https://github.com/conda-forge/staged-recipes"
	@echo "Subsequent pushes: need to for the python-seldon-core feedstock"
	@echo "Alternatively use anaconda upload to publish on own channel"

.PHONY: clean
clean:
	rm -rf seldon_core.egg-info
	rm -rf seldon_core/tensorflow
	rm -rf .eggs
	rm -rf .pytest_cache
	rm -rf dist
	rm -rf conda-bld
	rm -rf conda_pkg_path.txt
	rm -rf .empty
	rm -rf build
