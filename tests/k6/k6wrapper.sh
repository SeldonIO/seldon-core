#!/bin/sh

DIR="results"
METADATA="metadata.txt"
NOW=$(date +"%Y_%m_%d_%H_%M_%S")
UUID=$(uuidgen)
LABELS=$(cat /info/labels)
# extracts controller id from the labels: 
# eg: controller-uid="95a4c449-5cda-45a0-93e1-177caacc3639" job-name="k6"
JOBID=$(echo $LABELS | sed -n 's/.*controller-uid="\([a-zA-Z0-9-]\+\)".*/\1/p')
gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS
echo "start:"$(date) > $DIR/$METADATA
k6 $@
echo "end:"$(date) >> $DIR/$METADATA
echo "args:"$@ >> $DIR/$METADATA
echo "envs:"$(printenv) >> $DIR/$METADATA
echo "metadata:"$TEST_METADATA >> $DIR/$METADATA
echo "labels:"$LABELS >> $DIR/$METADATA 
gsutil cp -r $DIR gs://seldon-tmp/scv2-k6-results/${TEST_METADATA}_${JOBID}_${NOW}_${UUID}