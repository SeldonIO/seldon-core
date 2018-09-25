const express = require("express");
const app = express();
const bodyParser = require("body-parser");
const grpc = require("grpc");
const grpc_services = require("./prediction_grpc_pb");

module.exports = (
  user_model,
  api,
  port,
  rest_data_to_array,
  array_to_rest_data,
  array_to_grpc_data
) => {
  if (user_model.predict && typeof user_model.predict === "function") {
    console.log("Predict function loaded successfully");
  } else {
    console.log("Predict function not Found");
    process.exit(1);
  }
  let predict = user_model.predict.bind(user_model);

  if (api === "REST") {
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
        res.status(500).send(null);
      }
    });
    app.listen(port, () => {
      console.log(`NodeJs REST Microservice listening on port ${port}!`);
    });
  }

  if (api === "GRPC") {
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
