download_crts () {
    kubectl get secret -n kafka seldon-cluster-ca-cert -o jsonpath='{.data.ca\.p12}' | base64 -d > ca.p12
    kubectl get secret -n kafka seldon-cluster-ca-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
    kubectl get secret -n kafka seldon-cluster-ca-cert -o jsonpath='{.data.ca\.password}' | base64 -d > ca.password
    kubectl get secret -n kafka seldon-user -o jsonpath='{.data.user\.p12}' | base64 -d > user.p12
    kubectl get secret -n kafka seldon-user -o jsonpath='{.data.user\.password}' | base64 -d > user.password
    kubectl get svc -n kafka seldon-kafka-tls-bootstrap -o jsonpath='{.status.loadBalancer.ingress[0].ip}' > broker.ip
}

convert_ca_to_jks () {
    rm -f kafka-truststore.jks
    keytool -keystore kafka-truststore.jks -alias CARoot -import -file ca.crt -trustcacerts -keypass `cat ca.password` -storepass `cat ca.password` -noprompt
}

convert_user_to_jks () {
    rm -f kafka-keystore.jks
    keytool -importkeystore -srckeystore user.p12 -srcstoretype pkcs12 -destkeystore kafka-keystore.jks -deststoretype JKS -srcstorepass `cat user.password` -deststorepass `cat user.password`
}


create_config_properties () {
    cat config.properties.tmpl | sed s#TRUSTSTORE_PASSWORD#`cat ca.password`# | sed s#KEYSTORE_PASSWORD#`cat user.password`# | sed s#BROKER_IP#`cat broker.ip`# > ../config.properties
}

download_crts
convert_ca_to_jks
convert_user_to_jks
create_config_properties
