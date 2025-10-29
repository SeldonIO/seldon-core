# Build base conda image
make -C ../ docker-build-conda-base

# Build standard wrapper images
make -C ../ docker-build PYTHON_VERSION=3.12.12

# Push default tag image
make -C ../ docker-tag-base-python PYTHON_VERSION=3.12.12

# Build GPU images
make -C ../ docker-build-gpu PYTHON_VERSION=3.12.12
