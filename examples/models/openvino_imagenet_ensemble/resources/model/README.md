# OpenVINO prediction component

Model configuration is implemented using environment variables:

`XML_PATH`  - s3, gs or local path to xml file in OpenVINO model server

`BIN_PATH` - s3, gs or local path to bin file in OpenVINO model server

When using GCS make sure you added also GOOGLE_APPLICATION_CREDENTIALS variable and mounted correspondent token file.

In case is S3 or Minio storage add appropriate environment variables with the credentials.

Component is executing inference operation. Processing time is included in the components debug logs.

Model input and output tensors are determined automatically. There is assumed only one input tensor and output tensor.

### Building example:

```bash
s2i build -E environment_grpc . seldon_openvino_base:latest seldon-openvino-prediction:0.1
```
The base image `seldon_openvino_base:latest` should be created according to this [procedure](../../../../../wrappers/s2i/python_openvino)


### Local testing example:

```bash
docker run -it -v $GOOGLE_APPLICATION_CREDENTIALS:/etc/gcp.json -e GOOGLE_APPLICATION_CREDENTIALS=/etc/gcp.json \
 -e XML_PATH=gs://inference-eu/models_zoo/resnet_V1_50/resnet_V1_50.xml \
 -e BIN_PATH=gs://inference-eu/models_zoo/resnet_V1_50/resnet_V1_50.bin  
starting microservice
2019-02-05 11:13:32,045 - seldon_core.microservice:main:261 - INFO:  Starting microservice.py:main
2019-02-05 11:13:32,047 - seldon_core.microservice:main:292 - INFO:  Annotations: {}
path object /tmp/resnet_V1_50.xml
  net = IENetwork(model=xml_local_path, weights=bin_local_path)
2019-02-05 11:14:19,870 - seldon_core.microservice:main:354 - INFO:  Starting servers
2019-02-05 11:14:19,906 - seldon_core.microservice:grpc_prediction_server:333 - INFO:  GRPC microservice Running on port 5000
```



