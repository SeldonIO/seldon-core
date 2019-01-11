# NB: Tensorflow does not work with python 3.7 at present
# see https://github.com/tensorflow/tensorflow/issues/20444
make -C ../ build PYTHON_VERSION=3.7
