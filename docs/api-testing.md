# Testing Your Seldon Components

Whether you have wrapped your component using [our S2I wrappers](./wrappers/readme.md) or created your own wrapper you will want to test the Docker container standalone and also quickly within a running cluster. We have provided two python console scripts within the [seldon-core Python package](../python) to allow you to easily do this:

 * ```seldon-core-microservice-tester```
    * Allows you to test a docker component to check it respects the Seldon  internal microservice API.
 * ```seldon-core-api-tester```
    * Allows you to test the external endpoints for a running Seldon Deployment graph.

To use these, install the seldon-core package with ```pip install seldon-core```.

## Run your Wrapped Model

To test your model microservice you need to run it. If you have wrapped your model into a Docker container then you should run it and expose the ports. There are many examples in the notebooks in the [examples folders](https://github.com/SeldonIO/seldon-core/tree/master/examples/models) but essential if your model is wrapped in an image `myimage:0.1` then run:

```
docker run --name "my_model" -d --rm -p 5000:5000 myimage:0.1
```

Alternatively, if your component is a Python module you can run it directly from python using the core tool ```seldon-core-microservice``` (installed as part of the pip package `seldon-core`). This tool takes the name of the Python module as first argument and the API type REST or GRPC as second argument, for example if you have a file IrisClassifier.py in the current folder you could run:

```
seldon-core-microservice IrisClassifier REST
```

To get full details about this tool run `seldon-core-microservice --help`.

Next either use the [Microservce API tester](#microservice-api-tester) or testdirectly via [curl](#microservice-api-test-via-curl).

## Microservice API Tester

Use the ```seldon-core-microservice-tester``` script to test a packaged Docker microservice Seldon component.

```
usage: seldon-core-microservice-tester [-h] [--endpoint {predict,send-feedback}]
                          [-b BATCH_SIZE] [-n N_REQUESTS] [--grpc] [--fbs]
                          [-t] [-p]
                          contract host port

positional arguments:
  contract              File that contains the data contract
  host
  port

optional arguments:
  -h, --help            show this help message and exit
  --endpoint {predict,send-feedback}
  -b BATCH_SIZE, --batch-size BATCH_SIZE
  -n N_REQUESTS, --n-requests N_REQUESTS
  --grpc
  --fbs
  -t, --tensor
  -p, --prnt            Prints requests and responses
```

Example:

```
seldon-core-microservice-tester contract.json 0.0.0.0 5000 -p --grpc
```

The above sends a predict call to a gRPC component exposed at 0.0.0.0:5000 using the contract.json to create a random request.

You can find more examples in the [example models folder notebooks](../examples/models).

To understand the format of the contract.json see details [below](#api-contract).


## Microservice API Test via Curl
You can also test your component if run via Docker or from the command line via curl. An example for [Iris Classifier](http://localhost:8888/notebooks/sklearn_iris.ipynb) might be:

```
curl -g http://localhost:5000/predict --data-urlencode 'json={"data": {"names": ["sepal_length", "sepal_width", "petal_length", "petal_width"], "ndarray": [[7.233, 4.652, 7.39, 0.324]]}}'
```



# Seldon-Core API Tester for the External API 

Use the ```seldon-core-api-tester``` script to test a Seldon graph deployed to a kubernetes cluster.

```
usage: seldon-core-api-tester [-h] [--endpoint {predict,send-feedback}]
                              [-b BATCH_SIZE] [-n N_REQUESTS] [--grpc] [-t]
                              [-p] [--log-level {DEBUG,INFO,ERROR}]
                              [--namespace NAMESPACE]
                              [--oauth-port OAUTH_PORT]
                              [--oauth-key OAUTH_KEY]
                              [--oauth-secret OAUTH_SECRET]
                              contract host port [deployment]

positional arguments:
  contract              File that contains the data contract
  host
  port
  deployment

optional arguments:
  -h, --help            show this help message and exit
  --endpoint {predict,send-feedback}
  -b BATCH_SIZE, --batch-size BATCH_SIZE
  -n N_REQUESTS, --n-requests N_REQUESTS
  --grpc
  -t, --tensor
  -p, --prnt            Prints requests and responses
  --log-level {DEBUG,INFO,ERROR}
  --namespace NAMESPACE
  --oauth-port OAUTH_PORT
  --oauth-key OAUTH_KEY
  --oauth-secret OAUTH_SECRET

```

Example:

```
seldon-core-api-tester contract.json  0.0.0.0 8003 --oauth-key oauth-key --oauth-secret oauth-secret -p --grpc --oauth-port 8002 --endpoint send-feedback
```

 The above sends a gRPC send-feedback request to 0.0.0.0:8003 using the given oauth key/secret (assumes you are using the Seldon API Gateway) with the REST oauth-port at 8002 and use the contract.json file to create a random request. In this example you would have port-forwarded the Seldon api-server to local ports.

You can find more exampes in the [example models folder notebooks](../examples/models).

To understand the format of the contract.json see details [below](#api-contract).

## API Contract

Both tester scripts require you to provide a contract.json file defining the data you intend to send in a request and the response you expect back.

An example for the example Iris classification model is shown below:

```
{
    "features":[
	{
	    "name":"sepal_length",
	    "dtype":"FLOAT",
	    "ftype":"continuous",
	    "range":[4,8]
	},
	{
	    "name":"sepal_width",
	    "dtype":"FLOAT",
	    "ftype":"continuous",
	    "range":[2,5]
	},
	{
	    "name":"petal_length",
	    "dtype":"FLOAT",
	    "ftype":"continuous",
	    "range":[1,10]
	},
	{
	    "name":"petal_width",
	    "dtype":"FLOAT",
	    "ftype":"continuous",
	    "range":[0,3]
	}
    ],
    "targets":[
	{
	    "name":"class",
	    "dtype":"FLOAT",
	    "ftype":"continuous",
	    "range":[0,1],
	    "repeat":3
	}
    ]
}
```

Here we have 4 input features each of which is continuous in certain ranges. The response targets will be a repeated set of floats in the 0-1 range.

### Definition

There are two sections:

 * ```features``` : The types of the feature array that will be in the request
 * ```targets``` : The types of the feature array that will be in the response

Each section has a list of definitions. Each definition consists of:

  * ```name``` : String : The name of the feature
  * ```ftype``` : one of CONTINUOUS, CATEGORICAL : the type of the feature
  * ```dtype``` : One of FLOAT, INT : Required for ftype CONTINUOUS : What type of feature to create
  * ```values``` : list of Strings : Required for ftype CATEGORICAL : The possible categorical values
  * ```range``` : list of two numbers : Optional for ftype CONTINUOUS : The range of values (inclusive) that a continuous value can take
  * ```repeat``` : integer : Optional value for how many times to repeat this value
  * ```shape``` : array of integers : Optional value for the shape of array to coerce the values

