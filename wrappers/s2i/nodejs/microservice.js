const argparse = require("argparse");
const express = require("express");
const app = express();
const bodyParser = require("body-parser");
const nj = require("numjs");
app.use(bodyParser.urlencoded({ extended: true }));
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
      className[i] = "t" + i;
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

console.log(args.model, args.api, args.service, args.persistence);

if (args.service === "MODEL" && args.api === "REST") {
  app.post("/predict", async (req, res) => {
    try {
      body = JSON.parse(req.body.json);
      body = body.data;
    } catch (msg) {
      res.status(500).send("Cannot parse predict input json " + req.body.json);
    }
    if (predict && typeof predict === "function") {
      result = predict(rest_data_to_array(body), body.names);
      result = array_to_rest_data(result, body);
      res.status(200).send(result);
    } else {
      res.status(500).send(predict);
    }
  });
}

app.listen(port, async () => {
  predict = await loadModel(
    args.model,
    args.api,
    args.service,
    args.persistence
  );
  console.log("NodeJs Microservice listening on port 5000!");
});
