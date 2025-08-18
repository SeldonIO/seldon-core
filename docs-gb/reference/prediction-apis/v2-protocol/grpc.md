# gRPC

#### ServerLive

Check liveness of the inference server.

> rpc inference.GRPCInferenceService/ServerLive(ServerLiveRequest) returns ServerLiveResponse

#### ServerReady

Check readiness of the inference server.

> rpc inference.GRPCInferenceService/ServerReady(ServerReadyRequest) returns ServerReadyResponse

#### ModelReady

Check readiness of a model in the inference server.

> rpc inference.GRPCInferenceService/ModelReady(ModelReadyRequest) returns ModelReadyResponse

#### ServerMetadata

Get server metadata.

> rpc inference.GRPCInferenceService/ServerMetadata(ServerMetadataRequest) returns ServerMetadataResponse

#### ModelMetadata

Get model metadata.

> rpc inference.GRPCInferenceService/ModelMetadata(ModelMetadataRequest) returns ModelMetadataResponse

#### ModelInfer

Perform inference using a specific model.

> rpc inference.GRPCInferenceService/ModelInfer(ModelInferRequest) returns ModelInferResponse

#### RepositoryIndex

Get the index of model repository contents.

> rpc inference.GRPCInferenceService/RepositoryIndex(RepositoryIndexRequest) returns RepositoryIndexResponse

#### RepositoryModelLoad

Load or reload a model from a repository.

> rpc inference.GRPCInferenceService/RepositoryModelLoad(RepositoryModelLoadRequest) returns RepositoryModelLoadResponse

#### RepositoryModelUnload

Unload a model.

> rpc inference.GRPCInferenceService/RepositoryModelUnload(RepositoryModelUnloadRequest) returns RepositoryModelUnloadResponse

#### Messages

**InferParameter**

An inference parameter value.

| Field                                                                                                     | Type   | Description                |
| --------------------------------------------------------------------------------------------------------- | ------ | -------------------------- |
| [oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) parameter\_choice.bool\_param   | bool   | A boolean parameter value. |
| [oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) parameter\_choice.int64\_param  | int64  | An int64 parameter value.  |
| [oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) parameter\_choice.string\_param | string | A string parameter value.  |

**InferTensorContents**

The data contained in a tensor. For a given data type the tensor contents can be represented in "raw" bytes form or in the repeated type that matches the tensor's data type. Protobuf oneof is not used because oneofs cannot contain repeated fields.

| Field            | Type            | Description                                                                                                                                                                                                       |
| ---------------- | --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bool\_contents   | repeated bool   | Representation for BOOL data type. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements.                       |
| int\_contents    | repeated int32  | Representation for INT8, INT16, and INT32 data types. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements.    |
| int64\_contents  | repeated int64  | Representation for INT64 data types. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements.                     |
| uint\_contents   | repeated uint32 | Representation for UINT8, UINT16, and UINT32 data types. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements. |
| uint64\_contents | repeated uint64 | Representation for UINT64 data types. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements.                    |
| fp32\_contents   | repeated float  | Representation for FP32 data type. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements.                       |
| fp64\_contents   | repeated double | Representation for FP64 data type. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements.                       |
| bytes\_contents  | repeated bytes  | Representation for BYTES data type. The size must match what is expected by the tensor's shape. The contents must be the flattened, one-dimensional, row-major order of the tensor elements.                      |

**ModelInferRequest**

ModelInfer messages.

| Field                | Type                                                  | Description                                                                                                                                                                                                                                                                                                                                |
| -------------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| model\_name          | string                                                | The name of the model to use for inferencing.                                                                                                                                                                                                                                                                                              |
| model\_version       | string                                                | The version of the model to use for inference. If not given the server will choose a version based on the model and internal policy.                                                                                                                                                                                                       |
| id                   | string                                                | Optional identifier for the request. If specified will be returned in the response.                                                                                                                                                                                                                                                        |
| parameters           | map ModelInferRequest.ParametersEntry                 | Optional inference parameters.                                                                                                                                                                                                                                                                                                             |
| inputs               | repeated ModelInferRequest.InferInputTensor           | The input tensors for the inference.                                                                                                                                                                                                                                                                                                       |
| outputs              | repeated ModelInferRequest.InferRequestedOutputTensor | The requested output tensors for the inference. Optional, if not specified all outputs produced by the model will be returned.                                                                                                                                                                                                             |
| raw\_input\_contents | repeated bytes                                        | The data contained in an input tensor can be represented in "raw" bytes form or in the repeated type that matches the tensor's data type. Using the "raw" bytes form will typically allow higher performance due to the way protobuf allocation and reuse interacts with GRPC. For example, see https://github.com/grpc/grpc/issues/23231. |

To use the raw representation 'raw\_input\_contents' must be initialized with data for each tensor in the same order as 'inputs'. For each tensor, the size of this content must match what is expected by the tensor's shape and data type. The raw data must be the flattened, one-dimensional, row-major order of the tensor elements without any stride or padding between the elements. Note that the FP16 and BF16 data types must be represented as raw content as there is no specific data type for a 16-bit float type.

If this field is specified then InferInputTensor::contents must not be specified for any input tensor. |

**ModelInferRequest.InferInputTensor**

An input tensor for an inference request.

| Field      | Type                                                   | Description                                                                                                                               |
| ---------- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| name       | string                                                 | The tensor name.                                                                                                                          |
| datatype   | string                                                 | The tensor data type.                                                                                                                     |
| shape      | repeated int64                                         | The tensor shape.                                                                                                                         |
| parameters | map ModelInferRequest.InferInputTensor.ParametersEntry | Optional inference input tensor parameters.                                                                                               |
| contents   | InferTensorContents                                    | The input tensor data. This field must not be specified if tensor contents are being specified in ModelInferRequest.raw\_input\_contents. |

**ModelInferRequest.InferInputTensor.ParametersEntry**

| Field | Type           | Description |
| ----- | -------------- | ----------- |
| key   | string         | N/A         |
| value | InferParameter | N/A         |

**ModelInferRequest.InferRequestedOutputTensor**

An output tensor requested for an inference request.

| Field      | Type                                                             | Description                                  |
| ---------- | ---------------------------------------------------------------- | -------------------------------------------- |
| name       | string                                                           | The tensor name.                             |
| parameters | map ModelInferRequest.InferRequestedOutputTensor.ParametersEntry | Optional requested output tensor parameters. |

**ModelInferRequest.InferRequestedOutputTensor.ParametersEntry**

| Field | Type           | Description |
| ----- | -------------- | ----------- |
| key   | string         | N/A         |
| value | InferParameter | N/A         |

**ModelInferRequest.ParametersEntry**

| Field | Type           | Description |
| ----- | -------------- | ----------- |
| key   | string         | N/A         |
| value | InferParameter | N/A         |

**ModelInferResponse**

| Field                 | Type                                          | Description                                                                                                                                                                                                                                                                                                                                 |
| --------------------- | --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| model\_name           | string                                        | The name of the model used for inference.                                                                                                                                                                                                                                                                                                   |
| model\_version        | string                                        | The version of the model used for inference.                                                                                                                                                                                                                                                                                                |
| id                    | string                                        | The id of the inference request if one was specified.                                                                                                                                                                                                                                                                                       |
| parameters            | map ModelInferResponse.ParametersEntry        | Optional inference response parameters.                                                                                                                                                                                                                                                                                                     |
| outputs               | repeated ModelInferResponse.InferOutputTensor | The output tensors holding inference results.                                                                                                                                                                                                                                                                                               |
| raw\_output\_contents | repeated bytes                                | The data contained in an output tensor can be represented in "raw" bytes form or in the repeated type that matches the tensor's data type. Using the "raw" bytes form will typically allow higher performance due to the way protobuf allocation and reuse interacts with GRPC. For example, see https://github.com/grpc/grpc/issues/23231. |

To use the raw representation 'raw\_output\_contents' must be initialized with data for each tensor in the same order as 'outputs'. For each tensor, the size of this content must match what is expected by the tensor's shape and data type. The raw data must be the flattened, one-dimensional, row-major order of the tensor elements without any stride or padding between the elements. Note that the FP16 and BF16 data types must be represented as raw content as there is no specific data type for a 16-bit float type.

If this field is specified then InferOutputTensor::contents must not be specified for any output tensor. |

**ModelInferResponse.InferOutputTensor**

An output tensor returned for an inference request.

| Field      | Type                                                     | Description                                                                                                                                  |
| ---------- | -------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| name       | string                                                   | The tensor name.                                                                                                                             |
| datatype   | string                                                   | The tensor data type.                                                                                                                        |
| shape      | repeated int64                                           | The tensor shape.                                                                                                                            |
| parameters | map ModelInferResponse.InferOutputTensor.ParametersEntry | Optional output tensor parameters.                                                                                                           |
| contents   | InferTensorContents                                      | The output tensor data. This field must not be specified if tensor contents are being specified in ModelInferResponse.raw\_output\_contents. |

**ModelInferResponse.InferOutputTensor.ParametersEntry**

| Field | Type           | Description |
| ----- | -------------- | ----------- |
| key   | string         | N/A         |
| value | InferParameter | N/A         |

**ModelInferResponse.ParametersEntry**

| Field | Type           | Description |
| ----- | -------------- | ----------- |
| key   | string         | N/A         |
| value | InferParameter | N/A         |

**ModelMetadataRequest**

ModelMetadata messages.

| Field   | Type   | Description                                                                                                                            |
| ------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| name    | string | The name of the model.                                                                                                                 |
| version | string | The version of the model to check for readiness. If not given the server will choose a version based on the model and internal policy. |

**ModelMetadataResponse**

| Field      | Type                                          | Description                                                                                        |
| ---------- | --------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| name       | string                                        | The model name.                                                                                    |
| versions   | repeated string                               | The versions of the model available on the server.                                                 |
| platform   | string                                        | The model's platform. See Platforms.                                                               |
| inputs     | repeated ModelMetadataResponse.TensorMetadata | The model's inputs.                                                                                |
| outputs    | repeated ModelMetadataResponse.TensorMetadata | The model's outputs.                                                                               |
| parameters | map ModelMetadataResponse.ParametersEntry     | Optional default parameters for the request / response. NOTE: This is an extension to the standard |

**ModelMetadataResponse.ParametersEntry**

| Field | Type           | Description |
| ----- | -------------- | ----------- |
| key   | string         | N/A         |
| value | InferParameter | N/A         |

**ModelMetadataResponse.TensorMetadata**

Metadata for a tensor.

| Field      | Type                                                     | Description                                                                       |
| ---------- | -------------------------------------------------------- | --------------------------------------------------------------------------------- |
| name       | string                                                   | The tensor name.                                                                  |
| datatype   | string                                                   | The tensor data type.                                                             |
| shape      | repeated int64                                           | The tensor shape. A variable-size dimension is represented by a -1 value.         |
| parameters | map ModelMetadataResponse.TensorMetadata.ParametersEntry | Optional default parameters for input. NOTE: This is an extension to the standard |

**ModelMetadataResponse.TensorMetadata.ParametersEntry**

| Field | Type           | Description |
| ----- | -------------- | ----------- |
| key   | string         | N/A         |
| value | InferParameter | N/A         |

**ModelReadyRequest**

ModelReady messages.

| Field   | Type   | Description                                                                                                                            |
| ------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| name    | string | The name of the model to check for readiness.                                                                                          |
| version | string | The version of the model to check for readiness. If not given the server will choose a version based on the model and internal policy. |

**ModelReadyResponse**

| Field | Type | Description                                     |
| ----- | ---- | ----------------------------------------------- |
| ready | bool | True if the model is ready, false if not ready. |

**ModelRepositoryParameter**

An model repository parameter value.

| Field                                                                                                     | Type   | Description                |
| --------------------------------------------------------------------------------------------------------- | ------ | -------------------------- |
| [oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) parameter\_choice.bool\_param   | bool   | A boolean parameter value. |
| [oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) parameter\_choice.int64\_param  | int64  | An int64 parameter value.  |
| [oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) parameter\_choice.string\_param | string | A string parameter value.  |
| [oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) parameter\_choice.bytes\_param  | bytes  | A bytes parameter value.   |

**RepositoryIndexRequest**

| Field            | Type   | Description                                                                      |
| ---------------- | ------ | -------------------------------------------------------------------------------- |
| repository\_name | string | The name of the repository. If empty the index is returned for all repositories. |
| ready            | bool   | If true return only models currently ready for inferencing.                      |

**RepositoryIndexResponse**

| Field  | Type                                        | Description                    |
| ------ | ------------------------------------------- | ------------------------------ |
| models | repeated RepositoryIndexResponse.ModelIndex | An index entry for each model. |

**RepositoryIndexResponse.ModelIndex**

Index entry for a model.

| Field   | Type   | Description                                               |
| ------- | ------ | --------------------------------------------------------- |
| name    | string | The name of the model.                                    |
| version | string | The version of the model.                                 |
| state   | string | The state of the model.                                   |
| reason  | string | The reason, if any, that the model is in the given state. |

**RepositoryModelLoadRequest**

| Field            | Type                                           | Description                                                                                |
| ---------------- | ---------------------------------------------- | ------------------------------------------------------------------------------------------ |
| repository\_name | string                                         | The name of the repository to load from. If empty the model is loaded from any repository. |
| model\_name      | string                                         | The name of the model to load, or reload.                                                  |
| parameters       | map RepositoryModelLoadRequest.ParametersEntry | Optional model repository request parameters.                                              |

**RepositoryModelLoadRequest.ParametersEntry**

| Field | Type                     | Description |
| ----- | ------------------------ | ----------- |
| key   | string                   | N/A         |
| value | ModelRepositoryParameter | N/A         |

**RepositoryModelLoadResponse**

**RepositoryModelUnloadRequest**

| Field            | Type                                             | Description                                                                                                       |
| ---------------- | ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------- |
| repository\_name | string                                           | The name of the repository from which the model was originally loaded. If empty the repository is not considered. |
| model\_name      | string                                           | The name of the model to unload.                                                                                  |
| parameters       | map RepositoryModelUnloadRequest.ParametersEntry | Optional model repository request parameters.                                                                     |

**RepositoryModelUnloadRequest.ParametersEntry**

| Field | Type                     | Description |
| ----- | ------------------------ | ----------- |
| key   | string                   | N/A         |
| value | ModelRepositoryParameter | N/A         |

**RepositoryModelUnloadResponse**

**ServerLiveRequest**

ServerLive messages.

**ServerLiveResponse**

| Field | Type | Description                                              |
| ----- | ---- | -------------------------------------------------------- |
| live  | bool | True if the inference server is live, false if not live. |

**ServerMetadataRequest**

ServerMetadata messages.

**ServerMetadataResponse**

| Field      | Type            | Description                             |
| ---------- | --------------- | --------------------------------------- |
| name       | string          | The server name.                        |
| version    | string          | The server version.                     |
| extensions | repeated string | The extensions supported by the server. |

**ServerReadyRequest**

ServerReady messages.

**ServerReadyResponse**

| Field | Type | Description                                                |
| ----- | ---- | ---------------------------------------------------------- |
| ready | bool | True if the inference server is ready, false if not ready. |

## Scalar Value Types

### double

| Notes | C++ Type | Java Type | Python Type |
| ----- | -------- | --------- | ----------- |
|       | double   | double    | float       |

### float

| Notes | C++ Type | Java Type | Python Type |
| ----- | -------- | --------- | ----------- |
|       | float    | float     | float       |

### int32

| Notes                                                                                                                                           | C++ Type | Java Type | Python Type |
| ----------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------- | ----------- |
| Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32    | int       | int         |

### int64

| Notes                                                                                                                                           | C++ Type | Java Type | Python Type |
| ----------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------- | ----------- |
| Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64    | long      | int/long    |

### uint32

| Notes                          | C++ Type | Java Type | Python Type |
| ------------------------------ | -------- | --------- | ----------- |
| Uses variable-length encoding. | uint32   | int       | int/long    |

### uint64

| Notes                          | C++ Type | Java Type | Python Type |
| ------------------------------ | -------- | --------- | ----------- |
| Uses variable-length encoding. | uint64   | long      | int/long    |

### sint32

| Notes                                                                                                                | C++ Type | Java Type | Python Type |
| -------------------------------------------------------------------------------------------------------------------- | -------- | --------- | ----------- |
| Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32    | int       | int         |

### sint64

| Notes                                                                                                                | C++ Type | Java Type | Python Type |
| -------------------------------------------------------------------------------------------------------------------- | -------- | --------- | ----------- |
| Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64    | long      | int/long    |

### fixed32

| Notes                                                                                | C++ Type | Java Type | Python Type |
| ------------------------------------------------------------------------------------ | -------- | --------- | ----------- |
| Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32   | int       | int         |

### fixed64

| Notes                                                                                 | C++ Type | Java Type | Python Type |
| ------------------------------------------------------------------------------------- | -------- | --------- | ----------- |
| Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64   | long      | int/long    |

### sfixed32

| Notes              | C++ Type | Java Type | Python Type |
| ------------------ | -------- | --------- | ----------- |
| Always four bytes. | int32    | int       | int         |

### sfixed64

| Notes               | C++ Type | Java Type | Python Type |
| ------------------- | -------- | --------- | ----------- |
| Always eight bytes. | int64    | long      | int/long    |

### bool

| Notes | C++ Type | Java Type | Python Type |
| ----- | -------- | --------- | ----------- |
|       | bool     | boolean   | boolean     |

### string

| Notes                                                           | C++ Type | Java Type | Python Type |
| --------------------------------------------------------------- | -------- | --------- | ----------- |
| A string must always contain UTF-8 encoded or 7-bit ASCII text. | string   | String    | str/unicode |

### bytes

| Notes                                        | C++ Type | Java Type  | Python Type |
| -------------------------------------------- | -------- | ---------- | ----------- |
| May contain any arbitrary sequence of bytes. | string   | ByteString | str         |
