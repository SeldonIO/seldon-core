# set -exv
rm -rf vendor
mkdir -p vendor
git clone https://github.com/tensorflow/tensorflow.git || git --git-dir ./tensorflow/.git pull
PROTOC_OPTS='-I tensorflow --go_out=plugins=grpc:vendor'
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/framework/*.proto"
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/example/*.proto"
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/lib/core/*.proto"
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/protobuf/{saver,meta_graph}.proto"

mv vendor/tensorflow/core/* vendor/github.com/tensorflow/tensorflow/tensorflow/go/core