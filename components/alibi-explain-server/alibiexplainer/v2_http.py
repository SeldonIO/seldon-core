import numpy as np
from typing import Dict, List, Union

_v2tymap: Dict[str, np.dtype] = {
    "BOOL": np.dtype("bool"),
    "UINT8": np.dtype("uint8"),
    "UINT16": np.dtype("uint16"),
    "UINT32": np.dtype("uint32"),
    "UINT64": np.dtype("uint64"),
    "INT8": np.dtype("int8"),
    "INT16": np.dtype("int16"),
    "INT32": np.dtype("int32"),
    "INT64": np.dtype("int64"),
    "FP16": np.dtype("float32"),
    "FP32": np.dtype("float32"),
    "FP64": np.dtype("float64"),
}

_nptymap = dict([(value, key) for key, value in _v2tymap.items()])
_nptymap[np.dtype("float32")] = "FP32"  # Ensure correct mapping for ambiguous type

def _create_np_from_v2(data: list, ty: str, shape: list) -> np.array:
    npty = _v2tymap[ty]
    arr = np.array(data, dtype=npty)
    arr.shape = tuple(shape)
    return arr

def _create_v2_from_np(arr: np.ndarray, name: str, ty: str) -> Dict:
    if arr.dtype in _nptymap:
        return {
            "name": name,
            "datatype": ty,
            "data": arr.flatten().tolist(),
            "shape": list(arr.shape),
        }
    else:
        raise ValueError(f"Unknown numpy type {arr.dtype}")

# Only handle single input/output payloads
class KFServingV2RequestHandler():

    def extract_request(self, request: Dict) -> List:
        inputs = request["inputs"][0]
        data_type = inputs["datatype"]
        shape = inputs["shape"]
        data = inputs["data"]
        arr = _create_np_from_v2(data, data_type, shape)
        return arr.tolist()

    def extract_response(self, request: Dict) -> List:
        inputs = request["outputs"][0]
        data_type = inputs["datatype"]
        shape = inputs["shape"]
        data = inputs["data"]
        arr = _create_np_from_v2(data, data_type, shape)
        return arr

    def create_request(self, arr: np.ndarray, name: str, ty: str) -> Dict:
        req = {}
        data = _create_v2_from_np(arr, name, ty)
        req["inputs"] = [data]
        return req

    def extract_name(self, request: Dict) -> str:
        return request["inputs"][0]["name"]

    def extract_type(self, request: Dict) -> str:
        return request["inputs"][0]["datatype"]