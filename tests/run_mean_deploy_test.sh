#!/bin/bash

make cm_create_deployment
until make cm_check_deployment_ready
do
    echo "Waiting for deployment to be ready"
    sleep 1
done
sleep 10
make api_post_test
make cm_delete_deployment
