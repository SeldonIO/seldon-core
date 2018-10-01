# Nodejs s2i wrapper for seldon-core

## Steps to build model-template-app as a seldon model component

### Requirements for any NodeJS model to be created as a seldon component

- An example model is created at seldon-core/wrappers/s2i/nodejs/test/model-template-app/MyModel.js

```js
let MyModel = function() {};

MyModel.prototype.init = async function() {
  // A mandatory init method for the class to load run-time dependancies
  this.model = "My Awesome model";
};

MyModel.prototype.predict = function(data, names) {
  //A mandatory predict function for the model predictions
  console.log("Predicting ...");
  return data;
};

module.exports = MyModel;
```

- Mandatory prototype methods of the model `init` and `predict`. Also remember to include any other dependancy required for the model file to run

- A package.json file to contain all the dependancies and meta data for the model

- s2i environment variables in .s2i/enviroment file

```
MODEL_NAME=MyModel.js
API_TYPE=REST
SERVICE_TYPE=MODEL
PERSISTENCE=0
```

### Building the template nodejs app to create the model

```
cd test/model-template-app
npm install
npm start
```

This runs the model training file creates the model file required for the prediction

### Building Docker Image

```
docker build -t seldon-core-s2i-nodejs .
```

This builds the base wrapper image needed for any nodejs model to be deployed on seldon

### Building s2i nodejs model Image

```
s2i build -E ./test/model-template-app/.s2i/environment test/model-template-app seldonio/seldon-core-s2i-nodejs seldon-core-template-model
```

This creates the actual nodejs model image as a seldon component

### Test curl command for predicting with the nodejs model

```
curl  -d 'json={"data":{"names":[],"tensor":{"shape":[1,10],"values":[0,0,1,1,5,6,7,8,4,3]}}}' http://0.0.0.0:5000/predict
```

This command can be utilized to test the internal API of the model component

### Testing the image

Make sure the current user can run npm commands.

```
make test
```

### GRPC code-generated Proto JS Files

This is the code pre-generated using protoc and the Node gRPC protoc plugin, and the generated code can be found in various `*_pb.js` files.
The creation of the grpc srevice assumes these files to be present.

```
cd ../../../proto/
npm install -g grpc-tools
grpc_tools_node_protoc --js_out=import_style=commonjs,binary:../wrappers/s2i/nodejs/ --grpc_out=../wrappers/s2i/nodejs --plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` prediction.proto
cd ../wrappers/s2i/nodejs/
```

### Test using GRPC client

```
npm i
node grpc_client.js
```
