const argparse = require("argparse");
const express = require("express");
const app = express();
const bodyParser = require("body-parser");
const nj = require("numjs");
const grpc = require("grpc");
const grpc_messages = require("./prediction_pb");
const grpc_services = require("./prediction_grpc_pb");
const process = require("process");
let predict = null;
const port = process.env.PREDICTIVE_UNIT_SERVICE_PORT || 5000;

const loadModel = async function(model) {
  model = "./model/" + model;
  try {
    const MyModel = require(model);
    console.log("Loading Model", model);
    let x = new MyModel();
    await x.init();
    return x.predict.bind(x);
  } catch (msg) {
    return msg;
  }
};

const get_predict_classNames = function(size) {
  let className = [];
  if (size) {
    for (let i = 0; i < size; i++) {
      className[i] = "t:" + i;
    }
  }
  return className;
};

const rest_data_to_array = function(data) {
  if (data["tensor"]) {
    features = nj
      .array(data["tensor"]["values"])
      .reshape(data["tensor"]["shape"]);
  } else if (data["ndarray"]) {
    features = nj.array(data["ndarray"]);
  } else {
    features = nj.array([]);
  }
  return features.tolist();
};

const array_to_rest_data = function(array, original_datadef) {
  array = nj.array(array);
  let data = { names: get_predict_classNames(array.shape[1]) };
  if (original_datadef["tensor"]) {
    data["tensor"] = {
      shape: array.shape.length > 1 ? array.shape : [],
      values: array.flatten().tolist()
    };
  } else if (original_datadef["ndarray"]) {
    data["ndarray"] = array.tolist();
  } else {
    data["ndarray"] = array.tolist();
  }
  return data;
};

const array_to_grpc_data = function(array, original_datadef) {
  array = nj.array(array);

  var defdata = new grpc_messages.DefaultData();
  defdata.setNamesList(get_predict_classNames(array.shape[1]));
  if (original_datadef["tensor"]) {
    var tensorData = new grpc_messages.Tensor();
    tensorData.setShapeList(array.shape.length > 1 ? array.shape : []);
    tensorData.setValuesList(array.flatten().tolist());
    defdata.setTensor(tensorData);
  } else if (original_datadef["ndarray"]) {
    datadef.setNdarray(array.tolist());
  } else {
    datadef.setNdarray(array.tolist());
  }
  var data = new grpc_messages.SeldonMessage();
  data.setData(defdata);
  return data;
};

const parser = new argparse.ArgumentParser({
  description: "Seldon-core nodejs microservice builder",
  addHelp: true
});
parser.addArgument("--model", {
  type: "string",
  help: "Name of the model File"
});
parser.addArgument("--api", {
  type: "string",
  help: "Endpoints to be exposed REST or GRPC"
});
parser.addArgument("--service", {
  type: "string",
  help: "Service type"
});
parser.addArgument("--persistence", {
  type: "int",
  defaultValue: 0,
  help: "Persistence present or not"
});
const args = parser.parseArgs();

const getPredictFunction = async () => {
  predict = await loadModel(
    args.model,
    args.api,
    args.service,
    args.persistence
  );
  if (predict && typeof predict === "function") {
    console.log("Predict function loaded successfully");
    createServer();
  } else {
    console.log("Predict function not Found ", predict);
    process.exit(1);
  }
};
console.log(args.model, args.api, args.service, args.persistence);
getPredictFunction();

const createServer = () => {
  if (args.service === "MODEL" && args.api === "REST") {
    app.use(bodyParser.urlencoded({ extended: true }));
    app.post("/predict", (req, res) => {
      try {
        body = JSON.parse(req.body.json);
        body = body.data;
      } catch (msg) {
        console.log(msg);
        res.status(500).send("Cannot parse predict input json " + req.body);
      }
      if (predict && typeof predict === "function") {
        result = predict(rest_data_to_array(body), body.names);
        result = { data: array_to_rest_data(result, body) };
        res.status(200).send(result);
      } else {
        console.log("Predict function not Found");
        res.status(500).send(predict);
      }
    });
    app.listen(port, () => {
      console.log(`NodeJs REST Microservice listening on port ${port}!`);
    });
  }

  if (args.service === "MODEL" && args.api === "GRPC") {
    function predictEndpoint(call, callback) {
      let data = call.request.getData();
      let body = { names: data.getNamesList() };

      if (data.hasTensor()) {
        data = data.getTensor();
        body["tensor"] = {
          shape: data.getShapeList(),
          values: data.getValuesList()
        };
      } else {
        body["ndarray"] = data.getNdarray();
      }
      result = predict(rest_data_to_array(body), body.names);
      callback(null, array_to_grpc_data(result, body));
    }
    var server = new grpc.Server();
    server.addService(grpc_services.ModelService, { predict: predictEndpoint });
    server.bind("0.0.0.0:" + port, grpc.ServerCredentials.createInsecure());
    server.start();
    console.log(`NodeJs GRPC Microservice listening on port ${port}!`);
  }
};
