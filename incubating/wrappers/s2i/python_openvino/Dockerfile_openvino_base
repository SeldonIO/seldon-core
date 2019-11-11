FROM intelpython/intelpython3_core as DEV
RUN apt-get update && apt-get install -y \
            autoconf \
            automake \
            build-essential \
            ca-certificates \
            curl \
            git \
            gstreamer1.0-plugins-base \
            libavcodec-dev \
            libavformat-dev \
            libboost-regex-dev \
            libcairo2-dev \
            libgfortran3 \
            libglib2.0-dev \
            libgstreamer1.0-0 \
            libgtk2.0-dev \
            libopenblas-dev \
            libpango1.0-dev \
            libpng-dev \
            libssl-dev \
            libswscale-dev \
            libtool \
            libusb-1.0-0-dev \
            pkg-config \
            unzip \
            vim \
            wget

RUN wget https://cmake.org/files/v3.14/cmake-3.14.3.tar.gz && \
    tar -xvzf cmake-3.14.3.tar.gz && \
    cd cmake-3.14.3/  && \
    ./configure && \
    make -j$(nproc) && \
    make install

RUN echo "deb http://ftp.us.debian.org/debian/ jessie main contrib non-free" >> /etc/apt/sources.list && \
    echo "deb-src http://ftp.us.debian.org/debian/ jessie main contrib non-free" >> /etc/apt/sources.list && \
    apt update && \
    apt-get install -y g++-4.9
ENV CXX=/usr/bin/g++-4.9
RUN pip install cython numpy
ARG DLDT_DIR=/dldt-2019_R1.0.1
RUN git clone --depth=1 -b 2019_R1.0.1 https://github.com/opencv/dldt.git ${DLDT_DIR} && \
    cd ${DLDT_DIR} && git submodule init && git submodule update --recursive && \
    rm -Rf .git && rm -Rf model-optimizer

WORKDIR ${DLDT_DIR}
RUN curl -L https://github.com/intel/mkl-dnn/releases/download/v0.18/mklml_lnx_2019.0.3.20190220.tgz | tar -xz
WORKDIR ${DLDT_DIR}/inference-engine/build
RUN cmake -DGEMM=MKL  -DMKLROOT=${DLDT_DIR}/mklml_lnx_2019.0.3.20190220 -DENABLE_MKL_DNN=ON -DTHREADING=OMP -DCMAKE_BUILD_TYPE=Release ..
RUN make -j$(nproc)
WORKDIR ${DLDT_DIR}/inference-engine/ie_bridges/python/build
RUN cmake -DInferenceEngine_DIR=${DLDT_DIR}/inference-engine/build -DPYTHON_EXECUTABLE=$(which python) -DPYTHON_LIBRARY=/opt/conda/lib/libpython3.6m.so -DPYTHON_INCLUDE_DIR=/opt/conda/include/python3.6m ${DLDT_DIR}/inference-engine/ie_bridges/python && \
    make -j$(nproc)


FROM intelpython/intelpython3_core as PROD

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

RUN apt-get update && apt-get install -y --no-install-recommends \
            curl \
            ca-certificates \
            build-essential \
            python3-setuptools \
            vim

COPY --from=DEV /dldt-2019_R1.0.1/inference-engine/bin/intel64/Release/lib/*.so /usr/local/lib/
COPY --from=DEV /dldt-2019_R1.0.1/inference-engine/ie_bridges/python/bin/intel64/Release/python_api/python3.6/openvino/ /usr/local/lib/openvino/
COPY --from=DEV /dldt-2019_R1.0.1/mklml_lnx_2019.0.3.20190220/lib/lib*.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib
ENV PYTHONPATH=/usr/local/lib

RUN conda create --name myenv -y
ENV PATH /opt/conda/envs/myenv/bin:$PATH
RUN conda install -y tensorflow opencv && conda clean -a -y
WORKDIR /microservice

RUN pip install jaeger-client==3.13.0 seldon-core
RUN pip install --upgrade setuptools

RUN sed -i "s/max_workers=10/max_workers=1/g" /opt/conda/lib/python3.6/site-packages/seldon_core/wrapper.py
# set the workers to 1 - it avoids conflicts in reusing the IE model object in parallel calls
COPY ./s2i/bin/ /s2i/bin

EXPOSE 5000
