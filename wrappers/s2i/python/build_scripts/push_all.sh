# Push base conda image
make -C ../ docker-push-conda-base

# Push standard wrapper images
make -C ../ push_to_dockerhub PYTHON_VERSION=3.10.15

# Push default tag image
make -C ../ docker-push-base-python PYTHON_VERSION=3.10.15

# Push GPU images
make -C ../ docker-push-gpu PYTHON_VERSION=3.10.15
