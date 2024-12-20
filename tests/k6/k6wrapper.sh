#!/bin/sh

DIR="results"
NOW=$(date +"%Y_%m_%d_%H_%M_%S")
K6_ENV=$(env | grep -v -e SELDON -e SIMPLEST -e KUBERNETES -e GIT_COMMIT -e GIT_BRANCH -e HOSTNAME -e SHLVL -e PATH -e TEST_ID -e CLICKHOUSE -e ENDPOINT -e HOME -e NAMESPACE -e TERM -e PWD | sort )
K6_ENV_ENCODED=$(echo "${K6_ENV}" | base64)
LOADTEST_NOTES_ENCODED=$(echo "${LOADTEST_NOTES}" | base64)
METADATA_FILE="metadata-${TEST_ID}-${NOW}.txt"
SUMMARY_FILE="summary-${TEST_ID}-${NOW}.json"
RAW_FILE="rawTestData-${TEST_ID}-${NOW}.json"

UUID=$(uuidgen)
LABELS=$(cat /info/labels)
if [ -f "/info/clustermeta/cluster.uuid" ]; then
CLUSTER_UUID=$(cat /info/clustermeta/cluster.uuid)
else
CLUSTER_UUID="unknown"
fi
# extracts controller id from the labels:
# eg: controller-uid="95a4c449-5cda-45a0-93e1-177caacc3639" job-name="k6"
JOBID=$(echo $LABELS | sed -n 's/.*controller-uid="\([a-zA-Z0-9-]\+\)".*/\1/p')
TEST_TYPE=$(echo $LABELS | sed -n 's/.*job-name="\([a-zA-Z0-9-]\+\)".*/\1/p')
TEST_TARGET=${TEST_TYPE/-k6}

DEFAULT_OUTPUT_TYPE=none
K6_ARGS=""
for i in "$@"; do :; done # after this i is the last argument
K6_SCRIPT="$i"
SCENARIO=$(basename -- "$K6_SCRIPT" .js)

ENV_ID=$(echo "${K6_ENV}" | sha256sum | cut -d ' ' -f 1)

if [ -z "$OUTPUT_TYPE" ]; then
    OUTPUT_TYPE=${DEFAULT_OUTPUT_TYPE}
fi

# k6 pre-test
echo "start:"$(date) > $DIR/$METADATA_FILE
OUTERR=""
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
        if [ -z "${K6_CLICKHOUSE_USER}" ] ||      \
           [ -z "${K6_CLICKHOUSE_PASSWORD}" ] ||  \
           [ -z "${K6_CLICKHOUSE_HOST}" ] ||      \
           [ -z "${K6_CLICKHOUSE_PORT}" ] ||      \
           [ -z "${K6_CLICKHOUSE_DB}" ]; then
                   echo "The clickhouse OUTPUT_TYPE requires you to define the following variables: K6_CLICKHOUSE_USER, K6_CLICKHOUSE_PASSWORD, K6_CLICKHOUSE_HOST, K6_CLICKHOUSE_PORT, and K6_CLICKHOUSE_DB, in order to construct a connection string (DSN). Not all of them are set."
            OUTERR="1"
            if [ "{OUTPUT_TYPE_STRICT}" == "true" ]; then
                exit 1
            fi
            K6_ARGS="--summary-export $DIR/${SUMMARY_FILE}"
            K6_ARGS="${K6_ARGS} --out csv=$DIR/${RAW_FILE}"
        fi
        if [ -z "${OUTERR}" ]; then
            K6_CLICKHOUSE_DSN="${K6_CLICKHOUSE_USER}:${K6_CLICKHOUSE_PASSWORD}@${K6_CLICKHOUSE_HOST}:${K6_CLICKHOUSE_PORT}/${K6_CLICKHOUSE_DB}"
            K6_ARGS="--summary-export $DIR/${SUMMARY_FILE}"
            K6_ARGS="${K6_ARGS} --out clickhouse=clickhouse://${K6_CLICKHOUSE_DSN}"

            # Insert loadtest metadata into Clickhouse
            clickhouse-client \
                --host ${K6_CLICKHOUSE_HOST}         \
                --port ${K6_CLICKHOUSE_PORT}         \
                --user ${K6_CLICKHOUSE_USER}         \
                --password ${K6_CLICKHOUSE_PASSWORD} \
                --database ${K6_CLICKHOUSE_DB}       \
                --query "INSERT INTO ${K6_CLICKHOUSE_DB}.run_metadata \
                   ( \
                        ClusterID,  \
                        RunID,      \
                        StartAt,    \
                        TestTarget, \
                        K6Script,   \
                        K6Env,      \
                        K6EnvSum,   \
                        Notes       \
                    ) VALUES \
                    ( \
                        toUUIDOrZero('${CLUSTER_UUID}'), \
                        '${JOBID}',                      \
                        now64(3, 'Europe/London'),       \
                        '${TEST_TARGET}',                \
                        '${SCENARIO}',                   \
                        '${K6_ENV_ENCODED}',             \
                        '${ENV_ID}',                     \
                        '${LOADTEST_NOTES_ENCODED}'      \
                    )"
        fi
        ;;
    *)
        K6_ARGS=""
        ;;
esac

K6_CMD=$1
shift

./k6 ${K6_CMD} --env RUN_ID=${JOBID} --env K6_TEST_TARGET=${TEST_TARGET} --env K6_LOADTEST_SCRIPT=${SCENARIO} --env K6_ENV="${K6_ENV}" --env K6_ENV_ID=${ENV_ID} ${K6_ARGS} $@

K6_SUMMARY=$(cat $DIR/${SUMMARY_FILE} | base64)
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
    "clickhouse")
        if [ -z "${OUTERR}" ]; then
            # Insert loadtest summary into Clickhouse
            clickhouse-client \
                --host ${K6_CLICKHOUSE_HOST}         \
                --port ${K6_CLICKHOUSE_PORT}         \
                --user ${K6_CLICKHOUSE_USER}         \
                --password ${K6_CLICKHOUSE_PASSWORD} \
                --database ${K6_CLICKHOUSE_DB}       \
                --query "INSERT INTO ${K6_CLICKHOUSE_DB}.run_summary \
                   ( \
                        ClusterID,  \
                        RunID,      \
                        Summary,    \
                    ) VALUES \
                    ( \
                        toUUIDOrZero('${CLUSTER_UUID}'), \
                        '${JOBID}',                      \
                        '${K6_SUMMARY}'                  \
                    )"
        fi
        ;;
esac
