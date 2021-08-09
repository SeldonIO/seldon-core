# Push base conda image
make -C ../ push_to_dockerhub_conda_base

# Push standard wrapper images
make -C ../ push_to_dockerhub PYTHON_VERSION=3.6
make -C ../ push_to_dockerhub PYTHON_VERSION=3.7.10
make -C ../ push_to_dockerhub PYTHON_VERSION=3.8.10

# Push default tag image
make -C ../ push_to_dockerhub_base_python PYTHON_VERSION=3.7.10

# Push GPU images
make -C ../ push_gpu_to_dockerhub PYTHON_VERSION=3.6
make -C ../ push_gpu_to_dockerhub PYTHON_VERSION=3.7.10
make -C ../ push_gpu_to_dockerhub PYTHON_VERSION=3.8.10
