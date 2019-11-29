
DEST_REPO="tensorflow"
REPO="https://github.com/tensorflow/tensorflow.git"
if [ -d "${DEST_REPO}" ]; then
    git -C ${DEST_REPO} pull || true
else
    git clone ${REPO} ${DEST_REPO}
fi
mkdir -p vendor
PROTOC_OPTS='-I tensorflow --go_out=plugins=grpc:vendor'
eval "protoc $PROTOC_OPTS ${DEST_REPO}/tensorflow/core/framework/*.proto"
eval "protoc $PROTOC_OPTS ${DEST_REPO}/tensorflow/core/example/*.proto"
eval "protoc $PROTOC_OPTS ${DEST_REPO}/tensorflow/core/lib/core/*.proto"
eval "protoc $PROTOC_OPTS ${DEST_REPO}/tensorflow/core/protobuf/{saver,meta_graph}.proto"

