#!/bin/bash

protocol=${ENVOY_SECURITY_PROTOCOL:-PLAINTEXT}
if [ $protocol == "SSL" ]; then
    echo "Starting envoy with TLS"
    mkdir -p /tmp/certs
    printf '%s' "${ENVOY_XDS_CLIENT_TLS_KEY}" > /tmp/certs/tls.key
    printf '%s' "${ENVOY_XDS_CLIENT_TLS_CRT}" > /tmp/certs/tls.crt
    printf '%s' "${ENVOY_XDS_SERVER_TLS_CA}" > /tmp/certs/ca.crt   
    /usr/local/bin/envoy -c /etc/envoy-tls.yaml
else
    echo "Starting envoy"
    /usr/local/bin/envoy -c /etc/envoy.yaml
fi
