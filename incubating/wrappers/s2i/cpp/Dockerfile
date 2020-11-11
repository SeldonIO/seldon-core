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

WORKDIR /microservice

EXPOSE 5000
