/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

var messages = require("./prediction_pb");
var services = require("./prediction_grpc_pb");

var grpc = require("grpc");

function main() {
  var client = new services.ModelClient(
    "localhost:5000",
    grpc.credentials.createInsecure()
  );
  var tensorData = new messages.Tensor();
  tensorData.setShapeList([1, 10]);
  tensorData.setValuesList([0, 0, 1, 1, 5, 6, 7, 8, 4, 3]);

  var defdata = new messages.DefaultData();
  defdata.setTensor(tensorData);
  defdata.setNamesList([]);

  var request = new messages.SeldonMessage();
  request.setData(defdata);
  client.predict(request, function(err, response) {
    if (err) {
      console.log(err);
    } else {
      console.log(
        "Seldon Message => \n\nNames: ",
        response.getData().getNamesList(),
        "\n\nShape: ",
        response
          .getData()
          .getTensor()
          .getShapeList(),
        "\n\nValues: ",
        response
          .getData()
          .getTensor()
          .getValuesList()
      );
    }
  });
}

main();
