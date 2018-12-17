# Testing Your Seldon Components

Whether you have wrapped your component using [our S2I wrappers](./wrappers/readme.md) or created your own wrapper you will want to test the Docker container standalone and also quickly within a running cluster. We have provided two python console scripts within the [seldon-core Python package](../python) to allow you to easily do this:

 * ```seldon-core-tester```
    * Allows you to test a docker component to check it respects the Seldon  internal microservice API.
 * ```seldon-core-api-tester```
    * Allows you to test the external endpoints for a running Seldon Deployment graph.

To use these, install the seldon-core package with ```pip install seldon-core```.

## Seldon-Core API Tester

Use the ```seldon-core-api-tester``` script to test a Seldon graph deployed to a cluster.

```
usage: seldon-core-api-tester [-h] [--endpoint {predict,send-feedback}]
                              [-b BATCH_SIZE] [-n N_REQUESTS] [--grpc] [-t]
                              [-p] [--oauth-port OAUTH_PORT]
                              [--oauth-key OAUTH_KEY]
                              [--oauth-secret OAUTH_SECRET]
                              [--ambassador-path AMBASSADOR_PATH]
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
  -t, --tensor
  -p, --prnt            Prints requests and responses
  --oauth-port OAUTH_PORT
  --oauth-key OAUTH_KEY
  --oauth-secret OAUTH_SECRET
  --ambassador-path AMBASSADOR_PATH

```

Example:

```
seldon-core-api-tester contract.json  0.0.0.0 8003 --oauth-key oauth-key --oauth-secret oauth-secret -p --grpc --oauth-port 8002 --endpoint send-feedback
```

 The above sends a gRPC send-feedback request to 0.0.0.0:8003 using the given oauth key/secret (assumes you are using the Seldon API Gateway) with the REST oauth-port at 8002 and use the contract.json file to create a random request. In this example you would have port-forwarded the Seldon api-server to local ports.

You can find more exampes in the [example models folder notebooks](../examples/models).

## Microservice API Tester

Use the ```seldon-core-tester``` script to test a packaged Docker microservice Seldon component.

```
usage: seldon-core-tester [-h] [--endpoint {predict,send-feedback}]
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
seldon-core-tester contract.json 0.0.0.0 5000 -p --grpc
```

The above sends a predict call to a gRPC component exposed at 0.0.0.0:5000 using the contract.json to create a random request.

You can find more examples in the [example models folder notebooks](../examples/models).

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
