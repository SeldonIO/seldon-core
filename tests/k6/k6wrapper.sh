#!/bin/sh

DIR="results"
NOW=$(date +"%Y_%m_%d_%H_%M_%S")
METADATA_FILE="metadata-${TEST_ID}-${NOW}.txt"
SUMMARY_FILE="summary-${TEST_ID}-${NOW}.json"
RAW_FILE="rawTestData-${TEST_ID}-${NOW}.json"

UUID=$(uuidgen)
LABELS=$(cat /info/labels)
# extracts controller id from the labels:
# eg: controller-uid="95a4c449-5cda-45a0-93e1-177caacc3639" job-name="k6"
JOBID=$(echo $LABELS | sed -n 's/.*controller-uid="\([a-zA-Z0-9-]\+\)".*/\1/p')

DEFAULT_OUTPUT_TYPE=none
K6_ARGS=""

if [ -z "$OUTPUT_TYPE" ]; then
    OUTPUT_TYPE=${DEFAULT_OUTPUT_TYPE}
fi

# k6 pre-test
echo "start:"$(date) > $DIR/$METADATA_FILE
case $OUTPUT_TYPE in
    "csv")
        K6_ARGS="--summary-export $DIR/${SUMMARY_FILE}"
        K6_ARGS="${K6_ARGS} --out csv=$DIR/${RAW_FILE}"
        ;;
    "gs_bucket")
        if [ -f "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
            gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS
        fi
        K6_ARGS="--summary-export $DIR/${SUMMARY_FILE}"
        K6_ARGS="${K6_ARGS} --out csv=$DIR/${RAW_FILE}"
        ;;
    "clickhouse")
        if [ -z "${K6_CLICKHOUSE_DSN}" ]; then
            echo "Dataset Name (DSN) is not set in K6_CLICKHOUSE_DSN. The format needs to be clickhouse://user:pass@host:port/[DBName]"
            exit 1
        fi
        K6_ARGS="--summary-export $DIR/${SUMMARY_FILE}"
        K6_ARGS="${K6_ARGS} --out clickhouse=${K6_CLICKHOUSE_DSN}"
        ;;
    *)
        K6_ARGS=""
        ;;
esac

K6_CMD=$1
shift

./k6 ${K6_CMD} ${K6_ARGS} $@

# k6 post-test
echo "end:"$(date) >> $DIR/$METADATA_FILE
echo "args:"$@ >> $DIR/$METADATA_FILE
echo -e "envs:"$(printenv) >> $DIR/$METADATA_FILE
echo "metadata:"$TEST_METADATA >> $DIR/$METADATA_FILE
echo "labels:"$LABELS >> $DIR/$METADATA_FILE
case $OUTPUT_TYPE in
    "gs_bucket")
        if [ -f "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
            gsutil cp -r $DIR ${GS_BUCKET_NAME}/${TEST_METADATA}_${JOBID}_${NOW}_${UUID}
        fi
        exit 0
        ;;
esac
