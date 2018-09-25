const argparse = require("argparse");
const nj = require("numjs");
const grpc_messages = require("./prediction_pb");
const process = require("process");
const port = process.env.PREDICTIVE_UNIT_SERVICE_PORT || 5000;
let user_model = null;

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
  let features = null;
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

const dataFunctions = [
  rest_data_to_array,
  array_to_rest_data,
  array_to_grpc_data,
  get_predict_classNames
];

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

const loadModel = async function(model) {
  model = "./model/" + model;
  try {
    const MyModel = require(model);
    console.log("Loading Model", model);
    let x = new MyModel();
    await x.init();
    return x;
  } catch (msg) {
    return msg;
  }
};

const createServer = () => {
  if (args.service === "MODEL") {
    require("./model_microservice.js")(
      user_model,
      args.api,
      port,
      ...dataFunctions
    );
  }
  if (args.service === "TRANSFORMER") {
    require("./transformer_microservice.js")(
      user_model,
      args.api,
      port,
      ...dataFunctions
    );
  }
};

const getModelFunction = async () => {
  user_model = await loadModel(
    args.model,
    args.api,
    args.service,
    args.persistence
  );
  if (user_model) {
    console.log("Model Class loaded successfully");
    createServer();
  } else {
    console.log("Model Class could not be loaded ", user_model);
    process.exit(1);
  }
};

getModelFunction();
console.log(args.model, args.api, args.service, args.persistence);
