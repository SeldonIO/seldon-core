# Build base conda image
make -C ../ build_conda_base

# Build standard wrapper images
# NOTE: Conda 23.1.0 is the last version to support Python 3.7
make -C ../ build PYTHON_VERSION=3.7.10 CONDA_VERSION=23.1.0
make -C ../ build PYTHON_VERSION=3.8.10

# Tag the default image
make -C ../ tag_base_python PYTHON_VERSION=3.8.10

# Build GPU images
make -C ../ build_gpu PYTHON_VERSION=3.7.10 CONDA_VERSION=23.1.0
make -C ../ build_gpu PYTHON_VERSION=3.8.10
