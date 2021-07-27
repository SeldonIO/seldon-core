#!/bin/bash
echo "Before assembling"

/s2i/bin/assemble
rc=$?

if [ $rc -eq 0 ]; then
    echo "After successful assembling"
else
    echo "After failed assembling"
    exit $rc
fi

mkdir -p /tmp/.s2i/
cp image_metadata.json /tmp/.s2i/image_metadata.json
