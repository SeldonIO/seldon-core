ARG BASE_IMAGE_TAG
ARG BASE_IMAGE_NAME
FROM $BASE_IMAGE_NAME:$BASE_IMAGE_TAG

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

# Download cmake dependency for install of user module
RUN dnf install -y cmake

# Install pybind11 dependency for user module build
RUN git clone https://github.com/pybind/pybind11.git && \
        mkdir build && \
        cmake pybind11/ -Bbuild && \
        make -C build install && \
        rm -rf build pybind11

# Install C++ proto dependencies
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0-rc2/protobuf-all-3.14.0-rc-2.tar.gz && \
    tar -zxvf protobuf-all-3.14.0-rc-2.tar.gz && \
    cd protobuf-all-3.14.0-rc-2/ && \
    ./configure && \
    make && \
    make install 

WORKDIR /microservice

EXPOSE 5000
