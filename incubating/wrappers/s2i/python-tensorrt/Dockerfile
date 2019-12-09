FROM seldonio/seldon-core-s2i-python3:0.2

RUN apt-get update && apt-get install -y \
    	build-essential libssl1.0-dev libopencv-dev libopencv-core-dev python-pil \
	software-properties-common autoconf automake libtool pkg-config

WORKDIR /home

RUN git clone --single-branch -b change_ld_flags https://github.com/cliveseldon/dl-inference-server.git

RUN pip install --no-cache-dir --upgrade setuptools grpcio-tools

RUN cd dl-inference-server && \
    make -j4 -f Makefile.clients all pip

RUN pip install --no-cache-dir --upgrade dl-inference-server/build/dist/dist/inference_server-0.5.0-cp36-cp36m-linux_x86_64.whl

RUN rm -rf dl-inference-server/build

WORKDIR /microservice

