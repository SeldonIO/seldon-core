FROM seldonio/seldon-core-s2i-python3:0.2

RUN apt-get update && apt-get install -y \
	build-essential cmake clang-3.9 clang-format-3.9 \
	git curl zlib1g zlib1g-dev libtinfo-dev unzip autoconf automake libtool \
        clang-3.9 clang-format-3.9 \
        git \
        wget patch diffutils 

RUN apt-get clean autoclean && \
    apt-get autoremove -y
#RUN pip install --upgrade pip

WORKDIR /home

# Change to particular branch when stable
#RUN git clone --branch v0.7.0 https://github.com/NervanaSystems/ngraph.git && \

RUN git clone https://github.com/NervanaSystems/ngraph.git && \
    cd ngraph && \
    mkdir build && cd build && \
    cmake ../ -DNGRAPH_USE_PREBUILT_LLVM=TRUE -DCMAKE_INSTALL_PREFIX=/home/ngraph_dist -DNGRAPH_ONNX_IMPORT_ENABLE=TRUE && \
    make -j 6 && \
    make install && \
    cd .. && rm -rf build

RUN cd ngraph/python && \
    git clone --recursive -b allow-nonconstructible-holders https://github.com/jagerman/pybind11.git && \
    export PYBIND_HEADERS_PATH=$PWD/pybind11 && \
    export NGRAPH_CPP_BUILD_PATH=/home/ngraph_dist && \
    python3 setup.py bdist_wheel && \
    pip install -U dist/ngraph-0.7.0-cp36-cp36m-linux_x86_64.whl && \
    rm -rf build && rm -rf dist

RUN pip install git+https://github.com/NervanaSystems/ngraph-onnx/

WORKDIR /microservice

