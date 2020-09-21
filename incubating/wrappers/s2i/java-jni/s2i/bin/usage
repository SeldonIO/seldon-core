#!/bin/bash -e
cat <<EOF
This is the seldon-core-s2i-java-build S2I image:
To use it, install S2I: https://github.com/openshift/source-to-image

To create a template application clone https://github.com/seldonio/seldon-core.git and copy the appropriate folder for your needs from wrappers/s2i/java/test

Sample MODEL invocation:
------------------------

s2i build https://github.com/seldonio/seldon-core.git --context-dir=wrappers/s2i/python/test/model-template-app seldonio/seldon-core-s2i-java-build h2o-test:0.1 --runtime-image seldonio/seldon-core-s2i-java-runtime

You can then run the resulting image via:
docker run -p 5000:5000 seldon-core-template-model

And test:
curl  -d 'json={"data":{"ndarray":[[1.0,2.0]]}}' http://0.0.0.0:5000/predict


EOF
