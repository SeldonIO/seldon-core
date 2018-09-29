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
  if (
    user_model.transform_input &&
    typeof user_model.transform_input === "function"
  ) {
    console.log("Transform function loaded successfully");
  } else if (
    user_model.transform_output &&
    typeof user_model.transform_output === "function"
  ) {
    console.log("Transform function loaded successfully");
  } else {
    console.log(user_model);
    process.exit(1);
  }
  let transform_input = user_model.transform_input
    ? user_model.transform_input.bind(user_model)
    : null;
  let transform_output = user_model.transform_output
    ? user_model.transform_output.bind(user_model)
    : null;

  if (api === "REST") {
    app.use(bodyParser.urlencoded({ extended: true }));
    app.post("/transform-input", (req, res) => {
      try {
        body = JSON.parse(req.body.json);
        body = body.data;
      } catch (msg) {
        console.log(msg);
        res.status(500).send("Cannot parse transform input json " + req.body);
      }
      if (transform_input && typeof transform_input === "function") {
        result = transform_input(rest_data_to_array(body), body.names);
        result = { data: array_to_rest_data(result, body) };
        res.status(200).send(result);
      } else {
        console.log("Transform function not Found");
        res.status(500).send(null);
      }
    });
    app.post("/transform-output", (req, res) => {
      try {
        body = JSON.parse(req.body.json);
        body = body.data;
      } catch (msg) {
        console.log(msg);
        res.status(500).send("Cannot parse transform input json " + req.body);
      }
      if (transform_output && typeof transform_output === "function") {
        result = transform_output(rest_data_to_array(body), body.names);
        result = { data: array_to_rest_data(result, body) };
        res.status(200).send(result);
      } else {
        console.log("Transform function not Found");
        res.status(500).send(null);
      }
    });
    var server = app.listen(port, () => {
      console.log(`NodeJs REST Microservice listening on port ${port}!`);
    });
    function stopServer(code) {
      server.close();
      console.log(`About to exit with code: ${code}`);
    }
    process.on("SIGINT", stopServer.bind(this));
    process.on("SIGTERM", stopServer.bind(this));
  }

  if (api === "GRPC") {
    function inputEndpoint(call, callback) {
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
      result = transform_input(rest_data_to_array(body), body.names);
      callback(null, array_to_grpc_data(result, body));
    }
    function outputEndpoint(call, callback) {
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
      result = transform_output(rest_data_to_array(body), body.names);
      callback(null, array_to_grpc_data(result, body));
    }
    var server = new grpc.Server();
    server.addService(grpc_services.TransformerService, {
      transformInput: inputEndpoint
    });
    server.addService(grpc_services.OutputTransformerService, {
      transformOutput: outputEndpoint
    });
    server.bind("0.0.0.0:" + port, grpc.ServerCredentials.createInsecure());
    server.start();
    console.log(`NodeJs GRPC Microservice listening on port ${port}!`);
    function stopServer(code) {
      server.forceShutdown();
      console.log(`About to exit with code: ${code}`);
    }
    process.on("SIGINT", stopServer.bind(this));
    process.on("SIGTERM", stopServer.bind(this));
  }
};
