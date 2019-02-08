## Transformer component

Exemplary implementation of data transformation tasks.

Input transformation function accepts as input the binary representation of jpeg content.
It performs the following operations:
- convert compressed jpeg content to numpy array (BGR format)
- crop/resize the image to the square shape set in the environment variable `SIZE` (by default 224)
- transpose the data to NCWH


Output transformation function is consuming the imagenet classification models.
It is converting the array including probability for each imagenet classes into the class name.
It is returning 'human readable' string with most likely class name.
The function is using `CLASSES` environment variable to define the expected number of classes in the output. 
Depending on the model it could be 1000 (the default value) or 1001.


### Building example:
```bash
s2i build -E environment_grpc . seldon_openvino_base:latest seldonio/imagenet_transformer:0.1
```

The base image `seldon_openvino_base:latest` should be created according to this [procedure](../../../../../wrappers/s2i/python_openvino)