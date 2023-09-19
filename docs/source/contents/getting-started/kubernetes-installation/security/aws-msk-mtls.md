# AWS MSK mTLS

Seldon will run with [AWS MSK](https://aws.amazon.com/msk/).

At present we support mTLS authentication to MSK which can be run from a Kubernetes cluster inside or outside Amazon. If running outside your MSK cluster must have a public endpoint.

## Considerations

### Public Access to MSK Cluster

If you running your Kubernetes cluster outside AWS you will need to create a [public accessible MSK cluster](https://docs.aws.amazon.com/msk/latest/developerguide/public-access.html).

You will need to setup Kafka ACLs for your user where the username is the CommonName of the certificate of the client and allow full topic access. For example, to add a user with CN=myname to have full operations using the kafka-acls script with [mTLS config setup as described in AWS MSK docs](https://docs.aws.amazon.com/msk/latest/developerguide/msk-authentication.html):

```
kafka-acls.sh --bootstrap-server <mTLS endpoint>  --add --allow-principal User:CN=myname --operation All --topic '*' --command-config client.properties
```

You will also need to allow the connecting user to be able to perform admin tasks on the cluster so we can create topics on demand.

```
kafka-acls.sh --bootstrap-server <nTLS endpoint>  --add --allow-principal User:CN=myname --operation All --cluster '*' --command-config client.properties
```

You will need to allow group access also.

```
kafka-acls.sh --bootstrap-server <mTKS endpoint>  --add --allow-principal User:CN=myname --operation All --group '*' --command-config client.properties
```


## Create TLS Kubernetes Secrets

Create a secret for the client certificate you created. If you followed the [AWS MSK mTLS guide](https://docs.aws.amazon.com/msk/latest/developerguide/msk-authentication.html) you will need to export your private key from the JKS keystore. The certificate and chain will be provided in PEM format when you get the certificate signed. You can use these to create a secret with:

  * tls.key : PEM formatted private key
  * tls.crt : PEM formatted certificate
  * ca.crt : Certificate chain

```
kubectl create secret generic aws-msk-client --from-file=./tls.key --from-file=./tls.crt --from-file=./ca.crt -n seldon-mesh
```

Create a secret for the broker certificate. If following the [AWS MSK mTLS guide](https://docs.aws.amazon.com/msk/latest/developerguide/msk-authentication.html) you will need to export the trusstore of Amazon into PEM format and save as ca.crt.

To extract certificates from truststore do:

```
keytool -importkeystore -srckeystore truststore.jks    -destkeystore truststore.p12    -srcstoretype jks    -deststoretype pkcs12
openssl pkcs12 -in truststore.p12  -nodes -out trust.pem
cat trust.pem | sed  -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > ca.crt
```

Add ca.crt to a secret.

```
kubectl create secret generic aws-msk-broker-ca --from-file=./ca.crt -n seldon-mesh
```



### Example Helm install

We provide a template you can extend in `k8s/samples/values-aws-msk-kafka-mtls.yaml.tmpl`:

```{literalinclude} ../../../../../../k8s/samples/values-aws-msk-kafka-mtls.yaml.tmpl
:language: yaml
```

Copy this and modify by adding your broker endpoints.

```
helm install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh -f k8s/samples/values-aws-msk-kafka-mtls.yaml --set kafka.bootstrap=<your aws msk broker endpoints>
```

## Troubleshooting

First [check AWS MSK troubleshooting](https://docs.aws.amazon.com/msk/latest/developerguide/troubleshooting.html).

### No messages are being produced to the topics

Set the kafka config map debug setting to "all". For Helm install you can set `kafka.debug=all`.

If you see an error from the producer in the Pipeline gateway complaining about not enough insync replicas then the replication factor Seldon is using is less than the cluster setting for `min.insync.replicas` which for a default AWS MSK cluster defaults to 2. Ensure this is equal to that of the cluster. This value can be set in the Helm chart with `kafka.topics.replicationFactor`.
