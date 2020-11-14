ARG BASE_IMAGE_TAG
ARG BASE_IMAGE_NAME

FROM $BASE_IMAGE_NAME:$BASE_IMAGE_TAG

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

# Download cmake dependency for install of user module
RUN dnf install -y cmake
RUN dnf install -y python3-devel

# Install pybind11 dependency for user module build
RUN git clone --depth 1 --branch v2.6.1 https://github.com/pybind/pybind11 && \
        mkdir build && \
        cmake pybind11/ -Bbuild && \
        make -C build/ -j16 && \
        make -C build/ install && \
        rm -rf build pybind11

# Install C++ proto dependencies
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0-rc2/protobuf-all-3.14.0-rc-2.tar.gz && \
    tar -zxvf protobuf-all-3.14.0-rc-2.tar.gz && \
    cd protobuf-3.14.0-rc-2/ && \
    ./configure && \
    make -j16 && \
    make install && \
    ldconfig

COPY . _cpp/
RUN rm -rf _cpp/build/ || echo "no build dir"
RUN cmake _cpp/. -B_cpp/build/ -DSELDON_OPT_BUILD_STATIC=0
RUN make -C _cpp/build/ -j16
RUN make -C _cpp/build/ install

WORKDIR /microservice

EXPOSE 5000
